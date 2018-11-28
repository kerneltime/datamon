package storage

import (
	"bytes"
	"context"
	"io/ioutil"
	"runtime"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/oneconcern/datamon/internal"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAWSStore(t testing.TB) (Store, func()) {
	t.Helper()

	bid := internal.RandStringBytesMaskImprSrc(15)
	bucket := aws.String(bid)

	minioConfig := &aws.Config{
		Credentials:      credentials.NewStaticCredentials("access-key", "secret-key-thing", ""),
		Region:           aws.String("us-west-2"),
		Endpoint:         aws.String("http://127.0.0.1:9000"),
		S3ForcePathStyle: aws.Bool(true),
	}
	sess, err := session.NewSession(minioConfig)
	if err != nil {
		t.Skipf("minio is not running")
		runtime.Goexit()
	}
	cl := s3.New(sess)
	_, err = cl.CreateBucket(&s3.CreateBucketInput{
		Bucket: bucket,
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String("us-west-2"),
		},
	})
	require.NoError(t, err)

	cleanup := func() {
		_, _ = cl.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: bucket,
		})
	}

	_, err = cl.ListBuckets(nil)
	require.NoError(t, err)
	// t.Log(out.Buckets)

	up := s3manager.NewUploader(sess)
	_, err = up.UploadWithContext(aws.BackgroundContext(), &s3manager.UploadInput{
		Body:   bytes.NewBufferString("this is the text"),
		Bucket: bucket,
		Key:    aws.String("sixteentons"),
	})
	require.NoError(t, err)

	_, err = up.UploadWithContext(aws.BackgroundContext(), &s3manager.UploadInput{
		Body:   bytes.NewBufferString("this is the text for another thing"),
		Bucket: bucket,
		Key:    aws.String("seventeentons"),
	})
	require.NoError(t, err)
	return NewSThree(Bucket(*bucket), AWSConfig(minioConfig)), cleanup
}

func setupLocalFSStore(t testing.TB) (Store, func()) {
	t.Helper()

	fs := afero.NewMemMapFs()
	f, err := fs.Create("sixteentons")
	require.NoError(t, err)
	_, err = f.WriteString("this is the text")
	require.NoError(t, err)
	f.Close()

	ff, err := fs.Create("seventeentons")
	require.NoError(t, err)
	_, err = ff.WriteString("this is the text for another thing")
	require.NoError(t, err)
	ff.Close()

	return NewLocalFS(fs), func() {}
}

func TestHas(t *testing.T) {
	bs, cleanupAWS := setupAWSStore(t)
	defer cleanupAWS()
	has(t, bs)
	bs, cleanupLocalFS := setupLocalFSStore(t)
	has(t, bs)
	defer cleanupLocalFS()
}

func has(t *testing.T, store Store) {
	has, err := store.Has(context.Background(), "sixteentons")
	require.NoError(t, err)
	require.True(t, has)

	has, err = store.Has(context.Background(), "seventeentons")
	require.NoError(t, err)
	require.True(t, has)

	has, err = store.Has(context.Background(), "fifteentons")
	require.NoError(t, err)
	require.False(t, has)
}

func TestGet(t *testing.T) {
	bs, cleanupAWS := setupAWSStore(t)
	defer cleanupAWS()
	get(t, bs)

	bs, cleanupLocalFS := setupLocalFSStore(t)
	defer cleanupLocalFS()
	get(t, bs)
}

func get(t *testing.T, store Store) {
	rdr, err := store.Get(context.Background(), "sixteentons")
	require.NoError(t, err)
	b, err := ioutil.ReadAll(rdr)
	require.NoError(t, err)
	assert.Equal(t, "this is the text", string(b))

	rdr, err = store.Get(context.Background(), "seventeentons")
	require.NoError(t, err)
	b, err = ioutil.ReadAll(rdr)
	require.NoError(t, err)
	assert.Equal(t, "this is the text for another thing", string(b))
}

func TestKeys(t *testing.T) {
	bs, cleanupAWS := setupAWSStore(t)
	defer cleanupAWS()
	keys(t, bs)

	bs, cleanupLocalFS := setupLocalFSStore(t)
	defer cleanupLocalFS()
	keys(t, bs)
}

func keys(t *testing.T, store Store) {
	keys, err := store.Keys(context.Background())
	require.NoError(t, err)
	require.Len(t, keys, 2)
}

func TestDelete(t *testing.T) {
	bs, cleanupAWS := setupAWSStore(t)
	defer cleanupAWS()
	delete(t, bs)

	bs, cleanupLocalFS := setupLocalFSStore(t)
	defer cleanupLocalFS()
	delete(t, bs)
}

func delete(t *testing.T, store Store) {
	require.NoError(t, store.Delete(context.Background(), "seventeentons"))
	k, _ := store.Keys(context.Background())
	assert.Len(t, k, 1)
}

func TestClear(t *testing.T) {
	bs, cleanupAWS := setupAWSStore(t)
	defer cleanupAWS()
	clear(t, bs)

	bs, cleanupLocalFS := setupLocalFSStore(t)
	defer cleanupLocalFS()
	clear(t, bs)
}

func clear(t *testing.T, store Store) {
	require.NoError(t, store.Clear(context.Background()))
	k, _ := store.Keys(context.Background())
	require.Empty(t, k)
}

func TestPut(t *testing.T) {
	bs, cleanupAWS := setupAWSStore(t)
	defer cleanupAWS()
	put(t, bs)

	bs, cleanupLocalFS := setupLocalFSStore(t)
	defer cleanupLocalFS()
	put(t, bs)
}

func put(t *testing.T, store Store) {

	content := bytes.NewBufferString("here we go once again")
	err := store.Put(context.Background(), "eighteentons", content)
	require.NoError(t, err)

	rdr, err := store.Get(context.Background(), "eighteentons")
	require.NoError(t, err)
	b, err := ioutil.ReadAll(rdr)
	require.NoError(t, err)
	require.NoError(t, rdr.Close())

	assert.Equal(t, "here we go once again", string(b))

	k, _ := store.Keys(context.Background())
	assert.Len(t, k, 3)
}
