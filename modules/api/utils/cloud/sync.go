package cloud

import (
	"context"
	"fmt"
	"net/http"
)

func RemoveSAWithContext(ctx context.Context, areaID uint64, opts ...SCRequestOption) (err error) {
	request := map[string]interface{}{"is_remove": true}
	if err = syncToCloudWithContext(ctx, areaID, request, opts...); err != nil {
		return
	}
	return
}
func RemoveSAUserWithContext(ctx context.Context, areaID uint64, userID int) (err error) {
	request := map[string]interface{}{"remove_sa_user_id": userID}
	if err = syncToCloudWithContext(ctx, areaID, request); err != nil {
		return
	}
	return
}
func UpdateAreaNameWithContext(ctx context.Context, areaID uint64, name string) {
	request := map[string]interface{}{"sa_area_name": name}
	syncToCloudWithContext(ctx, areaID, request)
}
func syncToCloudWithContext(ctx context.Context, areaID uint64, request map[string]interface{}, opts ...SCRequestOption) (err error) {
	// 更新用户和家庭关系
	path := fmt.Sprintf("areas/%d", areaID)
	_, err = DoWithContext(ctx, path, http.MethodPut, request, opts...)
	if err != nil {
		return
	}
	return
}
