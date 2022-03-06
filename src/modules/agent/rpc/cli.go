package rpc

import (
	"bufio"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/toolkits/pkg/net/gobrpc"
	"github.com/ugorji/go/codec"
	"io"
	"net"
	"net/rpc"
	"reflect"
	"time"
)

type RpcCli struct {
	Cli *gobrpc.RPCClient
	ServerAddr string
	logger log.Logger
}

func InitRpcCli(serverAddr string, logger log.Logger) *RpcCli  {
	r := &RpcCli{
		ServerAddr: serverAddr,
		logger: logger,
	}
	return r
}

// 如果clic存在就返回，如果不存在就new 一个，复用
func (r *RpcCli) GetCli() error  {
	if r.Cli != nil{
		return nil
	}
	conn, err := net.DialTimeout("tcp", r.ServerAddr, time.Second * 5)
	if err != nil {
		level.Error(r.logger).Log("msg", "dial_server_failed", "serverAddr", r.ServerAddr, "err", err)
		return err
	}

	// 用bufferio做io解析提速
	var bufConn = struct {
		io.Closer
		*bufio.Reader
		*bufio.Writer
	}{conn, bufio.NewReader(conn), bufio.NewWriter(conn)}

	var mh codec.MsgpackHandle
	mh.MapType = reflect.TypeOf(map[string]interface{}(nil))

	rpcCodec := codec.MsgpackSpecRpc.ClientCodec(bufConn, &mh)
	client := rpc.NewClientWithCodec(rpcCodec)
	r.Cli = gobrpc.NewRPCClient(r.ServerAddr, client, 5*time.Second)
	return nil





}