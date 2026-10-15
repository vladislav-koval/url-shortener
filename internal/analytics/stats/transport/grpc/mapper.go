package statsgrpc

import (
	"fmt"

	"github.com/vladislav-koval/url-shortener/api/gen/analytics/v1"
	"github.com/vladislav-koval/url-shortener/internal/analytics/stats/service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toGetLinkAnalyticsResponse(
	result service.GetLinkAnalyticsResult,
) *analyticsv1.GetLinkAnalyticsResponse {
	timeline := make([]*analyticsv1.TimelinePoint, 0, len(result.Timeline))

	for _, point := range result.Timeline {
		timeline = append(timeline, &analyticsv1.TimelinePoint{
			BucketStart: timestamppb.New(point.BucketStart),
			Clicks:      int64(point.Clicks),
		})
	}

	return &analyticsv1.GetLinkAnalyticsResponse{
		TotalClicks:      int64(result.TotalClicks),
		Timeline:         timeline,
		Countries:        toBreakdown(result.Countries),
		Cities:           toBreakdown(result.Cities),
		Devices:          toBreakdown(result.Devices),
		OperatingSystems: toBreakdown(result.OperatingSystems),
		Browsers:         toBreakdown(result.Browsers),
		Referrers:        toBreakdown(result.Referrers),
	}
}

func toBreakdown(b service.Breakdown) *analyticsv1.Breakdown {
	items := make([]*analyticsv1.BreakdownItem, 0, len(b.Items))

	for _, item := range b.Items {
		items = append(items, &analyticsv1.BreakdownItem{
			Type:   toProtoBreakdownItemType(item.Type),
			Value:  item.Value,
			Clicks: int64(item.Clicks),
		})
	}

	return &analyticsv1.Breakdown{
		Items: items,
	}
}

func toProtoBreakdownItemType(t service.BreakdownItemType) analyticsv1.BreakdownItemType {
	switch t {
	case service.BreakdownItemValue:
		return analyticsv1.BreakdownItemType_BREAKDOWN_ITEM_TYPE_VALUE

	case service.BreakdownItemOther:
		return analyticsv1.BreakdownItemType_BREAKDOWN_ITEM_TYPE_OTHER

	case service.BreakdownItemUnknown:
		return analyticsv1.BreakdownItemType_BREAKDOWN_ITEM_TYPE_UNKNOWN

	default:
		return analyticsv1.BreakdownItemType_BREAKDOWN_ITEM_TYPE_UNSPECIFIED
	}
}

func toServiceGranularity(
	g analyticsv1.Granularity,
) (service.Granularity, error) {
	switch g {
	case analyticsv1.Granularity_GRANULARITY_UNSPECIFIED,
		analyticsv1.Granularity_GRANULARITY_AUTO:
		return service.GranularityAuto, nil

	case analyticsv1.Granularity_GRANULARITY_5_MINUTES:
		return service.Granularity5Minutes, nil

	case analyticsv1.Granularity_GRANULARITY_15_MINUTES:
		return service.Granularity15Minutes, nil

	case analyticsv1.Granularity_GRANULARITY_30_MINUTES:
		return service.Granularity30Minutes, nil

	case analyticsv1.Granularity_GRANULARITY_1_HOUR:
		return service.Granularity1Hour, nil

	case analyticsv1.Granularity_GRANULARITY_6_HOURS:
		return service.Granularity6Hours, nil

	case analyticsv1.Granularity_GRANULARITY_12_HOURS:
		return service.Granularity12Hours, nil

	case analyticsv1.Granularity_GRANULARITY_1_DAY:
		return service.Granularity1Day, nil

	case analyticsv1.Granularity_GRANULARITY_7_DAYS:
		return service.Granularity7Days, nil

	default:
		return "", fmt.Errorf("unsupported granularity: %d", g)
	}
}
