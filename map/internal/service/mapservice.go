package service

import (
	"context"
	"github.com/go-kratos/kratos/v2/errors"
	pb "map/api/mapService"
	"map/internal/biz"
)

type MapServiceService struct {
	pb.UnimplementedMapServiceServer
	msbiz *biz.MapServiceBiz
}

func NewMapServiceService(msbiz *biz.MapServiceBiz) *MapServiceService {
	return &MapServiceService{
		msbiz: msbiz,
	}
}

func (s *MapServiceService) GetDrivingInfo(ctx context.Context, req *pb.GetDrivingInfoReq) (*pb.GetDrivingReply, error) {
	distance, duration, err := s.msbiz.GetDriverInfo(req.Origin, req.Destination)
	if err != nil {
		return nil, errors.New(200, "LBS_ERROR", "lbs api error")
	}
	return &pb.GetDrivingReply{
		Origin:      req.Origin,
		Destination: req.Destination,
		Distance:    distance,
		Duration:    duration,
	}, nil
}
