package client

import (
	analyticsv1 "github.com/vladislav-koval/url-shortener/api/gen/analytics/v1"
	"github.com/vladislav-koval/url-shortener/internal/shortener/stats/service"
)

func toProtoGranularity(g service.Granularity) analyticsv1.Granularity {
	switch g {
	case service.GranularityAuto:
		return analyticsv1.Granularity_GRANULARITY_AUTO
	case service.Granularity5Minutes:
		return analyticsv1.Granularity_GRANULARITY_5_MINUTES
	case service.Granularity15Minutes:
		return analyticsv1.Granularity_GRANULARITY_15_MINUTES
	case service.Granularity30Minutes:
		return analyticsv1.Granularity_GRANULARITY_30_MINUTES
	case service.Granularity1Hour:
		return analyticsv1.Granularity_GRANULARITY_1_HOUR
	case service.Granularity6Hours:
		return analyticsv1.Granularity_GRANULARITY_6_HOURS
	case service.Granularity12Hours:
		return analyticsv1.Granularity_GRANULARITY_12_HOURS
	case service.Granularity1Day:
		return analyticsv1.Granularity_GRANULARITY_1_DAY
	case service.Granularity7Days:
		return analyticsv1.Granularity_GRANULARITY_7_DAYS
	default:
		return analyticsv1.Granularity_GRANULARITY_UNSPECIFIED
	}
}

func toServiceBreakdown(b *analyticsv1.Breakdown) service.Breakdown {
	if b == nil {
		return service.Breakdown{}
	}

	items := make([]service.BreakdownItem, 0, len(b.Items))

	for _, item := range b.Items {
		if item == nil {
			continue
		}

		items = append(items, service.BreakdownItem{
			Type:   toServiceBreakdownItemType(item.GetType()),
			Value:  item.GetValue(),
			Clicks: int(item.GetClicks()),
		})
	}

	return service.Breakdown{
		Items: items,
	}
}

func toServiceBreakdownItemType(
	t analyticsv1.BreakdownItemType,
) service.BreakdownItemType {
	switch t {
	case analyticsv1.BreakdownItemType_BREAKDOWN_ITEM_TYPE_VALUE:
		return service.BreakdownItemValue

	case analyticsv1.BreakdownItemType_BREAKDOWN_ITEM_TYPE_OTHER:
		return service.BreakdownItemOther

	case analyticsv1.BreakdownItemType_BREAKDOWN_ITEM_TYPE_UNKNOWN:
		return service.BreakdownItemUnknown

	default:
		return ""
	}
}

func toServiceTimeline(
	timeline []*analyticsv1.TimelinePoint,
) []service.TimelinePoint {
	result := make([]service.TimelinePoint, 0, len(timeline))

	for _, point := range timeline {
		if point == nil {
			continue
		}

		result = append(result, service.TimelinePoint{
			BucketStart: point.GetBucketStart().AsTime(),
			Clicks:      int(point.GetClicks()),
		})
	}

	return result
}

func analyticsFromProto(analytics *analyticsv1.GetLinkAnalyticsResponse) service.GetLinkAnalyticsResult {
	return service.GetLinkAnalyticsResult{
		TotalClicks:      int(analytics.TotalClicks),
		Timeline:         toServiceTimeline(analytics.Timeline),
		Countries:        toServiceBreakdown(analytics.Countries),
		Cities:           toServiceBreakdown(analytics.Cities),
		Devices:          toServiceBreakdown(analytics.Devices),
		OperatingSystems: toServiceBreakdown(analytics.OperatingSystems),
		Browsers:         toServiceBreakdown(analytics.Browsers),
		Referrers:        toServiceBreakdown(analytics.Referrers),
	}
}
