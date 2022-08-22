// Copyright 2013 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yar

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net/rpc"
	"sync"
)

type HttpServerCodec struct {
	r io.Reader

	req    serverRequest
	packer Packager

	mutex   sync.Mutex // protects seq, pending
	seq     uint64
	pending map[uint64]int64
}

func NewHttpServerCodec(reqBody io.Reader) *HttpServerCodec {
	return &HttpServerCodec{
		r:       reqBody,
		pending: make(map[uint64]int64),
	}
}

func (c *HttpServerCodec) ReadRequestHeader(r *rpc.Request) error {
	c.req.Reset()
	packer, err := readPack(c.r, &c.req)
	if err != nil {
		return err
	}

	c.packer = packer
	r.ServiceMethod = c.req.Method
	c.mutex.Lock()
	c.seq++
	c.pending[c.seq] = c.req.Id
	c.req.Id = 0
	r.Seq = c.seq
	c.mutex.Unlock()

	return nil
}

func (c *HttpServerCodec) ReadRequestBody(x interface{}) error {
	if x == nil {
		return nil
	}
	if c.req.Params == nil {
		return errMissingParams
	}
	return c.packer.Unmarshal(*c.req.Params, &x)

	// DEL: modify by zhiguang @ 20170319
	// err := c.packer.Unmarshal(*c.req.Params, &x)
	// if err != nil {
	// 	logger.Error("yar http server unmarshal request, err: %v, body: %v", err, utils.HexDumpToString(*c.req.Params))
	// }
	// return err
}

func (c *HttpServerCodec) WriteResponse(r *rpc.Response, x interface{}) (data []byte, err error) {
	c.mutex.Lock()
	id, ok := c.pending[r.Seq]
	if !ok {
		c.mutex.Unlock()
		return nil, errors.New("invalid sequence number in response")
	}
	delete(c.pending, r.Seq)
	c.mutex.Unlock()

	resp := serverResponse{
		Id:     id,
		Error:  "",
		Result: nil,
		Output: "",
		Status: 0,
	}

	if r.Error == "" {
		resp.Result = &x
	} else {
		resp.Error = r.Error
	}

	b := new(bytes.Buffer)
	w := bufio.NewWriter(io.Writer(b))
	Id := (int32)(resp.Id)
	err = writePack(w, c.packer, Id, &resp)
	if err != nil {
		return nil, err
	}
	err = w.Flush()
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
