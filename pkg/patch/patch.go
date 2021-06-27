package patch

type Patch struct {
	Id          string   `json:"id"`
	Patch       string   `json:"patch"`
	Author      string   `json:"author"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Prereqs     []string `json:"prereqs"`
	Body        []byte   `json:"-"`
	Weight      int      `json:"weight"`
	Filename    string   `json:"filename"`
}

type ByWeight []*Patch

func (by ByWeight) Len() int {
	return len(by)
}

func (by ByWeight) Swap(i, j int) {
	by[i], by[j] = by[j], by[i]
}

func (by ByWeight) Less(i, j int) bool {
	return by[i].Weight > by[j].Weight
}
