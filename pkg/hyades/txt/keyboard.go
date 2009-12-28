package txt

type KeyMap string

// XXX: DvorakMap not tested for typos.
const (
	ColemakMap KeyMap = " !\"#$%&'()*+,-./0123456789Pp<=>?@ABCGKETHLYNUMJ:RQSDFIVWXOZ[\\]^_`abcgkethlynumj;rqsdfivwxoz{|}~"
	DvorakMap  KeyMap = " !Q#$%&q()*}w'e[0123456789ZzW]E{@ANIHDYUJGCVPMLSRXO:KF><BT?/\\=^\"`anihdyujgcvpmlsrxo;kf.,bt/_|+~"
	QwertyMap  KeyMap = " !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"
)

func (self KeyMap) Map(key int) int {
	if key-32 < len(self) {
		return int(self[key-32])
	}
	return key
}
