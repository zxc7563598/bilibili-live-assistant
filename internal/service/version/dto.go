package version

// VersionInfo 版本信息
type VersionInfo struct {
	Version     string // 当前运行版本
	Commit      string // 当前构建提交
	Latest      string // 远端最新版本，检查失败时为空
	NeedUpdate  bool   // 是否可更新
	CheckFailed bool   // 是否检查失败（检查不通时不视为可更新）
}
