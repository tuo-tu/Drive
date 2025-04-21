package data

import (
	"valuation/internal/biz"
)

type PriceRuleData struct {
	data *Data
}

func NewPriceRuleInterface(data *Data) biz.PriceRuleInterface {
	return &PriceRuleData{data: data}
}

// PriceRuleData 实现 PriceRuleInterface，curr 表示什么？
func (prd *PriceRuleData) GetRule(cityid uint, curr int) (*biz.PriceRule, error) {
	pr := &biz.PriceRule{}
	// "start_at <= ? AND end_at > ?" 表示当前时刻在某时间范围内，对应该范围内的规则
	result := prd.data.Mdb.Where("city_id=? AND start_at <= ? AND end_at > ?", cityid, curr, curr).First(pr)
	if result.Error != nil {
		return nil, result.Error
	}
	return pr, nil
}
