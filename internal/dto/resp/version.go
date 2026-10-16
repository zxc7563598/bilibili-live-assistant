package resp

// VersionInfoResp 版本信息响应
type VersionInfoResp struct {
	Version     string `json:"version" example:"v1.0.0"`
	Commit      string `json:"commit" example:"a1b2c3d"`
	Latest      string `json:"latest" example:"v1.1.0"`
	NeedUpdate  bool   `json:"needUpdate" example:"true"`
	CheckFailed bool   `json:"checkFailed" example:"false"`
}
