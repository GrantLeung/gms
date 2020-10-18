package client

import (
	"errors"
	"fmt"

	"github.com/akkagao/gms/codec"
	"github.com/akkagao/gms/connection"
	"github.com/akkagao/gms/discovery"
	"github.com/akkagao/gms/protocol"
	"github.com/akkagao/gms/selector"
)

type Client struct {
	discovery   discovery.IDiscovery
	selector    selector.ISelector
	connection  map[string]connection.IConnection
	messagePack protocol.IMessagePack
	codecType   codec.CodecType
}

/*
NewClient 初始化客户端
*/
func NewClient(discovery discovery.IDiscovery) (IClient, error) {
	client := &Client{
		discovery:   discovery,
		connection:  make(map[string]connection.IConnection),
		messagePack: protocol.NewMessagePack(),
		codecType:   codec.JSON,
	}
	server, err := discovery.GetServer()
	if err != nil {
		return nil, fmt.Errorf("NewClient error %v", err)
	}
	client.selector = selector.NewRandomSelect(server)
	return client, nil
}

func (c *Client) SetCodecType(codecType codec.CodecType) error {
	if codec := codec.GetCodec(c.codecType); codec == nil {
		return errors.New("unsupped codecType,only supped['JSON','Msgpack','Gob']")
	}
	c.codecType = codecType
	return nil
}

func (c *Client) Call(serviceFunc string, request interface{}, response interface{}) error {
	serverKey := c.selector.Select()
	if serverKey == "" {
		return errors.New("can't find server")
	}

	connection := c.getCachedConnection(serverKey)

	codec := codec.GetCodec(c.codecType)

	codecByte, err := codec.Encode(request)
	if err != nil {
		fmt.Println(err)
	}

	message := protocol.NewMessage([]byte(serviceFunc), codecByte)
	eb, err := c.messagePack.Pack(message)
	if err != nil {
		// 错误处理
		fmt.Println(err)
	}
	err = connection.Send(eb)
	if err != nil {
		return err
	}

	// 读取返回结果
	message, err = connection.Read()
	if err != nil {
		return err
	}

	codec.Decode(message.GetData(), response)
	return nil
}

func (c *Client) getCachedConnection(address string) connection.IConnection {
	if connection, ok := c.connection[address]; ok {
		return connection
	}
	connection := c.generateClient(address)
	c.connection[address] = connection
	return connection
}

func (c *Client) generateClient(address string) connection.IConnection {
	return connection.NewConnection(address)
}

func (c *Client) Close() {
	// todo
	panic("implement me")
}
