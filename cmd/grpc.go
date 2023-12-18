package cmd

/*
 *  grpc 指令启动一个grpc服务端，模拟rawdb服务
 *	可以不依赖rawdb服务测试arc-consumer的grpc转发
 *
 */

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/kiga-hub/arc-consumer/pkg/file"
	"github.com/kiga-hub/arc/protobuf/pb"
	"github.com/kiga-hub/arc/protocols"
	"github.com/kiga-hub/arc/utils"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// GrpcCmd .
var GrpcCmd = &cobra.Command{
	Use:   "grpc",
	Short: "run data grpc arc-consumer",
	Run:   grpcRun,
}

var grpcPort int32
var grpcDir string

func init() {
	GrpcCmd.Flags().Int32VarP(&grpcPort, "port", "p", 8080, "grpc 监听端口")
	GrpcCmd.Flags().StringVarP(&grpcDir, "dirent", "d", "", "文件保存目录") // 目录参数是空， 不保存文件
}

// ProtoStream -
type ProtoStream struct {
	Key   []byte
	Value []byte
}

// FrameData -
type FrameData struct {
	Grpcmessage chan ProtoStream
}

// FrameDataCallback -
func (t *FrameData) FrameDataCallback(request pb.FrameData_FrameDataCallbackServer) (err error) {
	ctx := request.Context()
	fmt.Println("Client Connected")
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Client Disconnected Context Done")
			return ctx.Err()
		default:
		}

		tem, err := request.Recv()
		if err != nil {
			fmt.Println("Client Disconnected Err:", err)
			return request.SendAndClose(&pb.FrameDataResponse{Successed: false})
		}
		message := &ProtoStream{
			Key:   tem.Key,
			Value: tem.Value,
		}

		t.Grpcmessage <- *message
	}
}

// 创建grpc服务
func createServer(grpcmessage chan ProtoStream) *grpc.Server {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
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
			if err := listen.Close(); err != nil {
				fmt.Println(err)
			}
		}()
		fmt.Println("grpc server start port: ", grpcPort)
		err = grpcserver.Serve(listen)
		if err != nil {
			fmt.Println(err)
			return
		}
	}()
	return grpcserver
}

// GrpcStat - 接收数据统计结构
type GrpcStat struct {
	pkts  int64
	bytes int64
	total int64
	errs  int64
	items int64
}

// 命令行回调函数
func grpcRun(cmd *cobra.Command, args []string) {

	// 异常退出捕获
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered", "recover", r)
			debug.PrintStack()
			os.Exit(1)
		}
	}()

	// 消息管道内存分配
	grpcmessage := make(chan ProtoStream, 1024)
	defer close(grpcmessage)

	// 创建grpc服务端
	grpcserver := createServer(grpcmessage)

	var err error
	var fServe file.Handler
	if grpcDir != "" {
		// 初始化保存文件结构
		if fServe, err = file.New(
			file.WithConfig(&file.Config{
				DurationMin: 1,
				Dirent:      grpcDir,
				Timeout:     3000})); err != nil {
			panic(err)
		}

		// 启动保存文件服务
		go fServe.Start(context.Background())
		defer fServe.Stop()
	}

	// 关闭服务
	defer grpcserver.Stop()

	grpcStat := new(sync.Map)            // 统计结构
	frame := protocols.NewDefaultFrame() // 包结构

	// 服务端获取数据
	for p := range grpcmessage {
		var pidx uint32
		var lpidx uint32

		// 解析出ID值
		var clientid uint64
		for _, b := range p.Key {
			clientid <<= 8
			clientid += uint64(b)
		}

		// 统计
		v, _ := grpcStat.LoadOrStore(clientid, &GrpcStat{})
		stat := v.(*GrpcStat)
		stat.items++
		stat.total += int64(len(p.Value))

		// 详细解析判断
		for {
			if len(p.Value) <= int(pidx) {
				if len(p.Value) < int(pidx) {
					utils.Hexdump(fmt.Sprintf("%012X Size Error Offset: [%d:], Pidx: %d", clientid, lpidx, pidx), p.Value[lpidx:])
				}
				break
			}

			// 获取大小
			size := binary.BigEndian.Uint32(p.Value[pidx+5:]) + 9
			if len(p.Value) < int(pidx)+int(size) {
				utils.Hexdump(fmt.Sprintf("%012X Size Error Offset: [%d:], Pidx: %d, Frame Size %d", clientid, lpidx, pidx, size), p.Value[lpidx:])
				break
			}

			// 解析frame
			if err := frame.Decode(p.Value[pidx : pidx+size]); err != nil {
				fmt.Println(err)
				continue
			}

			// 判断包是否正确
			if frame.End != 0xFD || !bytes.Equal(frame.Head[:], []byte{0xFC, 0xFC, 0xFC, 0xFC}) {
				rpidx := uint32(len(p.Value))
				if len(p.Value)-int(lpidx) > 2048*5 {
					rpidx = lpidx + (2048 * 5)
				}
				utils.Hexdump(fmt.Sprintf("%012X Frame Error Offset: [%d:%d]", clientid, lpidx, rpidx), p.Value[lpidx:rpidx])
				stat.errs++
			} else {
				stat.pkts++
				stat.bytes += int64(size)
			}

			// 判断是否保存文件
			if fServe != nil {
				if err := fServe.Write(clientid, frame); err != nil {
					fmt.Println(err)
				}
			}

			lpidx = pidx
			pidx += size
		}

		// 打印统计日志
		if stat.items > 0 {
			fmt.Printf("[%s] clientid: %012X, size: %d, pkts: %d, bytes: %d, err: %d, avg burst pkts: %.2f, avg burst bytes: %.2f\n",
				time.Now().Format("2006-01-02 15:04:05"),
				clientid,
				stat.total,
				stat.pkts, stat.bytes, stat.errs,
				float64(stat.pkts)/float64(stat.items),
				float64(stat.bytes)/float64(stat.items),
			)
			stat.total = 0
			stat.items = 0
			stat.pkts = 0
			stat.bytes = 0
			stat.errs = 0
		}
	}
}
