package appconfig

// configValue 读取应用配置值，配置项缺失时返回空字符串
func (s *Service) configValue(key string) string {
	val, _ := s.appConfigCache.Get(key)
	return val
}
