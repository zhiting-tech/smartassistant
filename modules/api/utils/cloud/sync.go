package cloud

import (
	"context"
	"fmt"
	"net/http"
)

func RemoveSAWithContext(ctx context.Context, areaID uint64, opts ...SCRequestOption) {
	request := map[string]interface{}{"is_remove": true}
	syncToCloudWithContext(ctx, areaID, request, opts...)
}
func RemoveSAUserWithContext(ctx context.Context, areaID uint64, userID int) {
	request := map[string]interface{}{"remove_sa_user_id": userID}
	syncToCloudWithContext(ctx, areaID, request)
}
func UpdateAreaNameWithContext(ctx context.Context, areaID uint64, name string) {
	request := map[string]interface{}{"sa_area_name": name}
	syncToCloudWithContext(ctx, areaID, request)
}
func syncToCloudWithContext(ctx context.Context, areaID uint64, request map[string]interface{}, opts ...SCRequestOption) {
	// 更新用户和家庭关系
	path := fmt.Sprintf("areas/%d", areaID)
	_, err := DoWithContext(ctx, path, http.MethodPut, request, opts...)
	if err != nil {
		return
	}
}
