package config


type Oss struct {
	Driver string `json:"driver" yaml:"driver"`
	Aliyun Aliyun `json:"aliyun" yaml:"aliyun"`
}

type Aliyun struct {
	RegionId        string `json:"region_id" yaml:"region_id"`
	AccessKeyId     string `json:"access_key_id" yaml:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret" yaml:"access_key_secret"`
	RoleArn         string `json:"role_arn" yaml:"role_arn"`
	RoleSessionName string `json:"role_session_name" yaml:"role_session_name"`
}

