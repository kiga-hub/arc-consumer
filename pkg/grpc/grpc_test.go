package grpc

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"runtime"
	"testing"
	"time"

	"github.com/kiga-hub/arc/protobuf/pb"
	"github.com/kiga-hub/arc/protocols"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// ProtoStream -
type ProtoStream struct {
	Key    []byte
	Value  []byte
	IsStop bool
}

// FrameData -
type FrameData struct {
	Grpcmessage chan ProtoStream
}

// FrameDataCallback -
func (t *FrameData) FrameDataCallback(request pb.FrameData_FrameDataCallbackServer) (err error) {
	ctx := request.Context()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		tem, err := request.Recv()
		if err != nil {
			fmt.Println(err)
			return request.SendAndClose(&pb.FrameDataResponse{Successed: false})
			// return err
		}
		message := &ProtoStream{
			Key:    tem.Key,
			Value:  tem.Value,
			IsStop: false,
		}

		t.Grpcmessage <- *message
	}
}

func createServer(grpcmessage chan ProtoStream) *grpc.Server {
	listen, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		return nil
	}

	var kaep = keepalive.EnforcementPolicy{
		PermitWithoutStream: true,
	}

	var kasp = keepalive.ServerParameters{
		Time:    10 * time.Second,
		Timeout: 3 * time.Second,
	}

	grpcserver := grpc.NewServer(
		grpc.InitialWindowSize(1<<30),
		grpc.InitialConnWindowSize(1<<30),
		grpc.KeepaliveParams(kasp),
		grpc.KeepaliveEnforcementPolicy(kaep),
		grpc.MaxRecvMsgSize(4<<30),
		grpc.MaxSendMsgSize(4<<30),
	)

	pb.RegisterFrameDataServer(grpcserver, &FrameData{Grpcmessage: grpcmessage})
	go func() {
		defer func() {
			listen.Close()
			fmt.Println("grpc server stop")
		}()
		err = grpcserver.Serve(listen)
		if err != nil {
			fmt.Println(err)
			return
		}
	}()
	return grpcserver
}

func TestGrpc(t *testing.T) {
	grpcmessage := make(chan ProtoStream, 1024)
	defer close(grpcmessage)

	grpcserver := createServer(grpcmessage)

	// 服务端获取数据
	go func() {
		for p := range grpcmessage {
			var pidx uint32
			var count uint32
			fmt.Println("|:> arc-consumer")
			for {
				if len(p.Value) <= int(pidx) {
					break
				}

				// 获取ID
				var clientid uint64
				for _, b := range p.Value[pidx+23 : pidx+29] {
					clientid <<= 8
					clientid += uint64(b)
				}

				// 获取序号
				seq := int64(binary.BigEndian.Uint64(p.Value[pidx+9:]))

				// 获取大小
				size := binary.BigEndian.Uint32(p.Value[pidx+5:]) + 9

				fmt.Printf("|clientid: %d, seq: %d, size: %d\n", clientid, seq, size)
				pidx += size
				count++
			}
			fmt.Printf("[arc-consumer frame %d]\n", count)
		}
		fmt.Println("arc-consumer exit")
	}()

	// 客户端发送测试数据
	fmt.Printf("goroutine: %d\n", runtime.NumGoroutine())
	fmt.Println("start send...")
	clientDataSend()

	// 关闭服务
	grpcserver.Stop()

	fmt.Println("close service...")
	time.Sleep(time.Second * 3)
}

func clientDataSend() {
	// load config
	SetDefaultConfig()

	viper.Set(KeyGRPCEnable, true)

	// 创建
	srv := New()

	// 连接
	go srv.Start(context.Background())

	fmt.Printf("goroutine: %d\n", runtime.NumGoroutine())

	ticker := time.NewTicker(time.Millisecond * 1000)
	defer ticker.Stop()
	var count int64
	for {
		<-ticker.C
		if count >= 30 {
			break
		}
		if err := srv.Write(1, "1", getFrame(count)); err != nil {
			fmt.Println(err)
		}
		count++
	}
	fmt.Printf("goroutine: %d\n", runtime.NumGoroutine())

	time.Sleep(time.Second * 3)

	srv.Stop()

	fmt.Println("close client...")
	time.Sleep(time.Second * 3)
}

func getFrame(seq int64) []byte {
	// 1. 准备数据段
	audio := make([]byte, 2048)
	for i := 0; i < 2048; i++ {
		audio[i] = byte(i)
	}
	sa := protocols.NewDefaultSegmentArc()
	sa.Data = append(sa.Data, audio...)
	if err := sa.Validate(); err != nil {
		fmt.Println(err)
	}

	// 2. 添加数据段到组
	g := protocols.NewDefaultDataGroup()
	g.AppendSegment(sa)
	if err := g.Validate(); err != nil {
		fmt.Println(err)
	}

	// 3. 添加组到包
	p := protocols.NewDefaultFrame()
	p.SetID(15)
	p.Timestamp = seq
	p.SetDataGroup(g)

	// 4. 打包二进制
	buf := make([]byte, p.Size+9)
	_, err := p.Encode(buf)
	if err != nil {
		fmt.Println("%w", err)
	}
	return buf
}
