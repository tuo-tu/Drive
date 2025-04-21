package service

import (
	"context"
	"github.com/go-kratos/kratos/v2/errors"
	"valuation/internal/biz"

	pb "valuation/api/valuation"
)

type ValuationService struct {
	pb.UnimplementedValuationServer
	// 引用业务对象
	vb *biz.ValuationBiz
}

func NewValuationService(vb *biz.ValuationBiz) *ValuationService {
	return &ValuationService{
		vb: vb,
	}
}

func (s *ValuationService) GetEstimatePrice(ctx context.Context, req *pb.GetEstimatePriceReq) (*pb.GetEstimatePriceReply, error) {
	// 得到距离、时长
	distance, duration, err := s.vb.GetDrivingInfo(ctx, req.Origin, req.Destination)
	if err != nil {
		return nil, errors.New(200, "MAP ERROR", "get driving info error")
	}
	// 费用
	price, err := s.vb.GetPrice(ctx, distance, duration, 1, 9)
	if err != nil {
		return nil, errors.New(200, "PRICE ERROR", "cal price error")
	}
	return &pb.GetEstimatePriceReply{
		Origin:      req.Origin,
		Destination: req.Destination,
		Price:       price,
	}, nil
}
