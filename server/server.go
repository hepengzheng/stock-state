package server

import (
	statepb "github.com/hepengzheng/stock-state/api/statebp"
)

type Server struct {
	statepb.UnimplementedStateServer
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) GetStock(stream statepb.State_GetStockServer) error {
	panic("implement me")
}
