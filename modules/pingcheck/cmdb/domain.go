package cmdb

import (
	"github.com/signmem/falcon-plus/modules/pingcheck/g"
)

var (
	CmdbHostRecord []HostInfo
)

func getExcludeDomains() (allDomains []string, allowDomains []string, extDomains []string) {

	// 过滤条件优先级如下
	// 1 excludedomain 最高级，匹配则放置  extdomains
	// 2 同时满足 excludetag 及 excludedeploytype 则放置 extdomains
	// 3 excludedeploytype == int 结构，需要先预先获取

	api := "/app/query"
	query := "name="

	cmdbData, err := CmdbApiQuery(api, query)

	if err != nil {
		g.Logger.Debugf("GetExcludeDomains() error: %s", err)
		return
	}

	//extStatus := false
	//deployTagStatus := false

	if len(cmdbData.Object) > 0 {

		for _, info := range cmdbData.Object {

			domain := info.Name
			var allow bool
			allow = true

			allDomains = append(allDomains, domain)

			if len(g.Config().ExcludeAppName) > 0 {
				for _, excString := range g.Config().ExcludeAppName {
					if excString == domain {
						if g.Config().Debug {
							g.Logger.Debugf("[EXT DOMAINS] %s Match Exclude App Name", domain)
						}
						extDomains = append(extDomains, domain)
						allow = false
						break
					}
				}
			}

			if allow == false {
				continue
			}

			for _, groupName := range g.Config().ExcludeBusGroup {
				if info.BusGroupName == groupName && allow == true {
					extDomains = append(extDomains, domain)
					allow = false
					break
				}
			}

			if allow == true {
				allowDomains = append(allowDomains, domain)
			}
		}
	}

	if g.Config().Debug {
		g.Logger.Debug("=====[CMDB APP INDEX START] ================")
		g.Logger.Debugf("[CMDB APP INDEX] 过滤应用 总共:  %d", len(extDomains))
		g.Logger.Debugf("[CMDB APP INDEX] 合法应用 总共:  %d", len(allowDomains))
		g.Logger.Debugf("[CMDB APP INDEX] 获取应用 总共:  %d", len(allDomains))
		g.Logger.Debug("=====[CMDB APP INDEX END] ================")
	}

	return
}

