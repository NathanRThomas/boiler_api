/*! \file file_cabnet.go
 *  \brief Wrapper around the aws class for uploading items of interest
 */

package toolz

import (
    "fmt"
    "os"
    "net/http"
    "image"
    "image/jpeg"
    "image/png"
    "io"
    "time"
    "io/ioutil"
    
	"github.com/nfnt/resize"
    "github.com/oliamb/cutter"
    )

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//
const aws_bucket_bot    = ""
const aws_cache_bot     = ".cloudfront.net"

const img_target_width		= 800

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type FileCabnet_c struct {
    aws AWS_c
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief handles the shared logic of uploading an image once we know what bucket it's going to
 */
func (this *FileCabnet_c) upload (bucket, cache, uniqueString, ext, fileLoc string) (url string, err error) {
    url, err = this.aws.UploadUnique(bucket, uniqueString, ext, fileLoc)
    if err == nil {
        if len(cache) < 1 {
            url = "https://s3.amazonaws.com/" + bucket + "/" + url     // the upload was successful, so let's return the full url
        } else {
            url = "https://" + cache + "/" + url
        }
    } else {
        err = ErrChk(fmt.Errorf("Upload : %s : %s : %s : %s : %s :: %s",
                                bucket, cache, uniqueString, ext, fileLoc, err.Error()))
    }
    os.Remove(fileLoc)
    return //and done
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Once we have a Reader for the image in question, we can do the rest from here
 */
func (this *FileCabnet_c) LocalUpload (file io.Reader, id string, compress, cropTop bool) (url string, err error) {
	//generates a unique name for this file, so we don't stomp
	locFile := "/tmp/" + this.aws.Hash(fmt.Sprintf("%s:%s", id, time.Now().UnixNano()), 32)
	
	out, err := os.Create(locFile)
	if err != nil {
		err = ErrChk(fmt.Errorf("fileUpload - Unable to create the file for writing. Check your write access privilege ::" + err.Error()))
		return
	}
	
	// write the content from POST to the file
	_, err = io.Copy(out, file)
	out.Close()	//we're done with this
	if err != nil {
		err = ErrChk(fmt.Errorf("fileUpload - unable to write to file ::", err.Error()))
		return
	}
	
	newFile, err := os.Open(locFile)
	if err == nil {
		defer newFile.Close()
	
		b, err := ioutil.ReadAll(newFile)
		ErrChk(err)
		ext := ""
		//fileType := http.DetectContentType(b)
		//fmt.Println(fileType)
		switch http.DetectContentType(b) {
			case "image/jpeg", "image/jpg", "text/plain; charset=utf-8":
				ext = "jpg"
			case "image/png":
				ext = "png"
			case "image/gif":
				ext = "gif"
				compress = false	//can't compress these
			case "mp4", "video/mp4":
				ext = "mp4"	//default to a video
				compress = false    //can't compress these
			default:
				return "", fmt.Errorf("Unsupported file type")
		}
		
		if err == nil {
			//we want to compress this image
			if compress {
				if ext == "png" {	//we need to make it a jpg
					err = this.ConvertPngToJpg (locFile)
					ext = "jpg"
				}
				
				if err == nil {
					targetHeight := 0
					if cropTop { targetHeight = img_target_width  / 2 }
					err = this.Compress (locFile, img_target_width, targetHeight, true)	//compress to 800 pixels wide
				}
			}
			
			if err == nil {
				url, err = this.upload(aws_bucket_bot, aws_cache_bot, this.aws.Hash(locFile, 10), ext, locFile)	//finally upload it
			}
		}
	} else {
		ErrChk(err)
	}
	return
}

/*! \brief Converts a file from png to jpg for us.  Makes it smaller
 */
func (this *FileCabnet_c) ConvertPngToJpg (in string) error {
    file, err := os.Open(in)        //open this file
    if err != nil { return ErrChk(err) }
    
    img, err := png.Decode(file)    //decode from a png to an image object
    if err != nil { return ErrChk(err) }
    
    file.Close()    //close this handle, we're going to overwrite the file
    out, err := os.Create(in)   //re-create the file
    jpeg.Encode(out, img, &jpeg.Options{Quality: 80})  //create a jpg
    out.Close()
    return ErrChk(err)  //we're good
}

/*! \brief Handles compressing and resizing an image
 */
func (this *FileCabnet_c) Compress (in string, targetWidth uint, targetHeight int, force bool) error {
    //open the existing image
    file, err := os.Open(in)
    if err != nil { return ErrChk(err) }
    
    //decode the image
	img, err := jpeg.Decode(file)
    if err != nil { return ErrChk(err) }
    file.Close()
    
    if force == false {
        //open it again to get the dimensions
        file, _ = os.Open(in)
        baseConfig, _, err := image.DecodeConfig(file)
        ErrChk(err)
        if baseConfig.Width < int(targetWidth) { return nil }    //nothing to do here
    }
	
	//resize the image
	m := resize.Resize(targetWidth, 0, img, resize.Lanczos3)
    
    //see if we need to crop from the top
    if targetHeight > 0 {
        m, err = cutter.Crop(m, cutter.Config{Width: int(targetWidth), Height: targetHeight})
    }
    
    //save it to disk
	out, err := os.Create(in)
    if err != nil { return ErrChk(err) }
	jpeg.Encode(out, m, nil)
	out.Close()
    return nil  //we're good
}
