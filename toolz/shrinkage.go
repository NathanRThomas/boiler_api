/*! \file shrinkage.go
    \brief Stuff shrinks when it's cold
*/

package toolz

import (
    "fmt"
    "bytes"
    "compress/gzip"
    "encoding/json"
    "io"
)

const defaultCompressionLevel = 5   //gzip default compression level

type Shrinkage_c struct { }

/*! \brief Compresses what we're caching.  This makes some things larger yes, but i think overall we're saving space
*/
func (this Shrinkage_c) Compress (val interface{}) ([]byte) {
    if jsVal, err := json.Marshal(val); err == nil {
        return this.Shrink(jsVal)
    } else {
        ErrChk(err)
    }
    return nil    
}

func (this Shrinkage_c) Shrink (val []byte) ([]byte) {
    var buffer bytes.Buffer
    if gz, err := gzip.NewWriterLevel (&buffer, defaultCompressionLevel); err == nil {
        if _, err = gz.Write(val); err == nil {
            if err = gz.Close(); err == nil {
                return buffer.Bytes()
            } else {
                ErrChk(err)
            }
        } else {
            ErrChk(err)
        }
    } else {
        ErrChk(err)
    }
    return nil  //this is bad
}

/*! \brief Uncompresses what we've compressed, cause ya, i think that makes sense
*/
func (this Shrinkage_c) Uncompress (data []byte, val interface{}) []byte {
    b := bytes.NewReader(data)
    out := new(bytes.Buffer)
    
    if reader, err := gzip.NewReader(b); err == nil {
        defer reader.Close()
        io.Copy(out, reader)
        if val != nil {
            ErrChk(json.Unmarshal(out.Bytes(), val))
        } else {
            return out.Bytes()
        }
    } else {
        ErrChk(fmt.Errorf(err.Error() + " :: " + string(data)))
    }
    return nil //this didn't work
}
