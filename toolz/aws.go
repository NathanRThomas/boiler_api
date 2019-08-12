/*! \file aws.go
 *  \brief Contains calls to aws
*/

package toolz

import (
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
    
    "io/ioutil"
    "crypto/sha256"
    "encoding/hex"
    "time"
    "bytes"
    "fmt"
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

const awsLocation = "us-east-1"

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type AWS_c struct {
    
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//


  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief I needed a way to create short hashes for index and uniqueness purposes
 *  This defaults to a lenght of 10 charactrs
 *  This defaults to epoch time unless another seed is passed
*/
func (this *AWS_c) Hash(seed string, length int) string {
    if length <= 0 { length = 10 }
    if len(seed) < 1 { seed = fmt.Sprintf("%d", time.Now().UnixNano()) }
    hash := sha256.Sum256([]byte(seed))
    return hex.EncodeToString(hash[:])[:length]
}


/*! \brief Uploads something to an aws bucket with a unique key
*/
func (this *AWS_c) UploadUnique (bucket, key, ext, fileLocation string) (string, error) {
    key = this.Hash("", 0) + "_" + key
    return this.UploadFile (bucket, key, ext, fileLocation)
}

/*! \brief When we need to specificy the file name already
 *  nothing special
 */
func (this *AWS_c) UploadFile (bucket, key, ext, fileLocation string) (string, error) {
    payload, err := ioutil.ReadFile(fileLocation)
    if err == nil {
        return this.Upload (bucket, key, ext, payload)
    } else {
        err = fmt.Errorf("UploadFile reading file :: %s", err.Error())
    }
    return "", err
}

func (this *AWS_c) Upload (bucket, key, ext string, payload []byte) (final string, err error) {
    svc := s3.New(session.New(), &aws.Config{Region: aws.String(awsLocation)})
    
    key = key + "." + ext
    contentType := "video/mp4"
    cacheControl := ""
    
    switch ext {
    case "jpeg", "jpg":
        contentType = "image/jpeg"
    case "png":
        contentType = "image/png"
    case "gif":
        contentType = "image/gif"
    case "js":
        contentType = "application/javascript"
        cacheControl = "max-age=0"
    }

    params := &s3.PutObjectInput{
        Bucket:             aws.String(bucket), // Required
        Key:                aws.String(key),  // Required
        ACL:                aws.String("public-read"),
        Body:               bytes.NewReader(payload),
        ContentLength:      aws.Int64(int64(len(payload))),
        ContentType:        aws.String(contentType),
        CacheControl:       aws.String(cacheControl),
        Metadata: map[string]*string{
            "Content-Type": aws.String(contentType),
        },
    }
    _, err = svc.PutObject(params)
    
    if err != nil {
        err = fmt.Errorf("Upload :: %s", err.Error())
    } else {    //this worked
        final = key //return what the actual key is
    }
    return
}

/*! \brief Deletes a file from a bucket
*/
func (this *AWS_c) Delete (bucket, key string) (error) {
    svc := s3.New(session.New(), &aws.Config{Region: aws.String(awsLocation)})

    params := &s3.DeleteObjectInput {
        Bucket:             aws.String(bucket), // Required
        Key:                aws.String(key),  // Required
    }

    _, err := svc.DeleteObject(params)
    return err
}
