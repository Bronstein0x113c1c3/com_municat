package networking

import "net"

func NewConnHTTP2(endpoint string) (net.Listener, error) {
	lis, err := net.Listen("tcp", endpoint)
	if err != nil {
		return nil, err
	}
	return lis, err
}
