package utils

import (
	"io"
	"log"
)

//Forward forward the stream1st to stream2nd and return the ch to receive the total bytes cnt
//blockSize and blockHandle to custome buf prefix process
func Forward(stream1st io.Reader, stream2nd io.Writer, ch chan int64, blockSize int, bufHandle func([]byte) []byte) error {
	log.Println("starting forward")
	bufLen := 32 * 1024
	var writeCnt int64
	writeCnt = 0
	if bufHandle != nil {
		bufLen = blockSize
	}
	buf := make([]byte, bufLen, bufLen)
	defer release(ch, writeCnt)
	for {
		n, err := stream1st.Read(buf)
		if n > 0 {
			nw, errw := stream2nd.Write(buf[0:n])
			if nw > 0 {
				writeCnt += int64(nw)
			}
			if errw != nil {
				//log.Printf("write error:%v", errw)
				return errw
			}
			if nw != n {
				return io.ErrShortWrite
			}
		}
		if err != nil {
			if err != io.EOF {
				//log.Printf("read error:%v", err)
				return err
			}
			return nil
		}
	}

}

func release(ch chan int64, x int64) {
	ch <- x
	log.Println("exit forward")
}

func Int64ToKB(a int64) float64 {
	x := float64(a)
	return x / 1024
}
