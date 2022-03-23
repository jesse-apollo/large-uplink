package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"cloud.google.com/go/storage"
	log "github.com/sirupsen/logrus"
)

// given a base64 encoded hmac, compare it
/*func validMAC(message, key []byte, messageMAC string) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	messageMACDecoded, _ := hex.DecodeString(messageMAC)
	return hmac.Equal(messageMACDecoded, expectedMAC)
}*/

// generate
func generateMAC(message, key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)

	hmac := mac.Sum(nil)

	return hex.EncodeToString(hmac)
}

func uploadGCSFile(input io.Reader, bucket, object, compositionID, signature string) error {

	log.Debugf("uploadGCSFile: %s/%s", bucket, object)
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	o := client.Bucket(bucket).Object(object)

	// Upload an object with storage.Writer.
	wc := o.NewWriter(ctx)
	if _, err = io.Copy(wc, input); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}

	objectAttrsToUpdate := storage.ObjectAttrsToUpdate{
		Metadata: map[string]string{
			"signature":      signature,
			"composition-id": compositionID,
		},
	}
	if _, err := o.Update(ctx, objectAttrsToUpdate); err != nil {
		return fmt.Errorf("ObjectHandle(%q).Update: %v", object, err)
	}

	log.Debugf("blob %v uploaded", object)
	return nil
}

func downloadGCSFile(bucket, object, signature string) ([]byte, string, error) {

	log.Debugf("downloadGCSFile: %s/%s (%s)", bucket, object, signature)

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	o := client.Bucket(bucket).Object(object)

	rc, err := o.NewReader(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("Object(%q).NewReader: %v", object, err)
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, "", fmt.Errorf("ioutil.ReadAll: %v", err)
	}

	log.Debugf("Blob %v downloaded.", object)

	// get metadata to confirm access
	attrs, err := o.Attrs(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("no attributes could be loaded: %v", err)
	}

	if attrs.Metadata["signature"] != signature {
		return nil, "", fmt.Errorf("signature is invalid: %s != %s", attrs.Metadata["signature"], signature)
	}

	return data, attrs.Metadata["composition-id"], nil
}

func deleteGCSFile(bucket, object string) error {
	log.Debugf("deleteGCSFile: %s/%s", bucket, object)

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	o := client.Bucket(bucket).Object(object)

	if err := o.Delete(ctx); err != nil {
		return fmt.Errorf("Object(%q).Delete: %v", object, err)
	}
	log.Debugf("Blob %v deleted", object)
	return nil
}
