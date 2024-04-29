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

			// fmt.Printf("domain %s, depart %s", info.Name, info.DeptFullname)

			domain := info.Name
			var allow bool
			allow = true

			allDomains = append(allDomains, domain)

			if len(g.Config().ExcludeDomains) > 0 {
				for _, excString := range g.Config().ExcludeDomains {
					if excString == domain {
						if g.Config().Debug {
							g.Logger.Debugf("domain Match ExcludeDomains: %s", domain)
						}
						extDomains = append(extDomains, domain)
						allow = false
						break
					}
				}
			}

			for _, groupName := range g.Config().ExcludeDomains {
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
		g.Logger.Debug("===== 额外域名 domain type 检测匹配: start  ================")
		g.Logger.Debugf("domain exclude 总共:  %d", len(extDomains))
		g.Logger.Debugf("domain allow 总共:  %d", len(allowDomains))
		g.Logger.Debugf("domain total 总共:  %d", len(allDomains))
		g.Logger.Debug("===== 额外域名 domain type 检测匹配: end  ================")
	}

	return
}

