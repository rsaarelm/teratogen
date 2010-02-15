package teratogen

import (
	"hyades/entity"
	"hyades/geom"
	"hyades/num"
)

const spawnsPerLevel = 32

//func DrawPos(pos geom.Pt2I) (screenX, screenY int) {
//	return TileW*pos.X + xDrawOffset, TileH*pos.Y + yDrawOffset
//}
//
//func CenterDrawPos(pos geom.Pt2I) (screenX, screenY int) {
//	return TileW*pos.X + xDrawOffset + TileW/2, TileH*pos.Y + yDrawOffset + TileH/2
//}
//
//func Draw(g gfx.Graphics, spriteId string, x, y int) {
//	sx, sy := DrawPos(geom.Pt2I{x, y})
//	DrawSprite(g, spriteId, sx, sy)
//}

/*
func (self *World) Draw(g gfx.Graphics) {
	self.drawTerrain(g)
	self.drawEntities(g)
}
*/

func Spawn(name string) *Blob {
	manager := GetManager()
	guid := assemblages[name].MakeEntity(manager)

	return GetBlobs().Get(guid).(*Blob)
}

func SpawnAt(name string, pos geom.Pt2I) (result *Blob) {
	result = Spawn(name)
	result.MoveAbs(pos)
	return
}

func SpawnRandomPos(name string) (result *Blob) {
	return SpawnAt(name, GetSpawnPos())
}

func clearNonplayerEntities() {
	// Bring over player object and player's inventory.
	player := GetPlayer()
	keep := make(map[entity.Id]bool)
	keep[player.GetGuid()] = true
	for ent := range player.RecursiveContents().Iter() {
		keep[ent.(*Blob).GetGuid()] = true
	}

	for o := range GetBlobs().EntityComponents().Iter() {
		pair := o.(*entity.IdComponent)
		if _, ok := keep[pair.Entity]; !ok {
			defer GetManager().RemoveEntity(pair.Entity)
		}
	}
}

func makeSpawnDistribution(depth int) num.WeightedDist {
	weightFn := func(item interface{}) float64 {
		proto := assemblages[item.(string)][BlobComponent].(*blobTemplate)
		return SpawnWeight(proto.Scarcity, proto.MinDepth, depth)
	}
	values := make([]interface{}, len(assemblages))
	i := 0
	for name, _ := range assemblages {
		values[i] = name
		i++
	}
	return num.MakeWeightedDist(weightFn, values)
}


// TODO: Move to SDL client
/*
func (self *World) drawEntities(g gfx.Graphics) {
	// Make a vector of the entities sorted in draw order.
	seq := new(vector.Vector)
	for o := range self.Entities().Iter() {
		ent := o.(*Blob)
		if ent.GetParent() != nil {
			// Skip entities inside something.
			continue
		}
		seq.Push(ent)
	}
	alg.PredicateSort(entityEarlierInDrawOrder, seq)

	for sorted := range seq.Iter() {
		e := sorted.(*Blob)
		pos := e.GetPos()
		seen := GetLos().Get(pos) == LosSeen
		mapped := seen || GetLos().Get(pos) == LosMapped
		// TODO: Draw static (item) entities from map memory.
		if mapped {
			if seen || !IsMobile(e) {
				//				Draw(g, e.IconId, pos.X, pos.Y)
			}
		}
	}
}

// TODO: Move to SDLClient
func entityEarlierInDrawOrder(i, j interface{}) bool {
	return i.(*Blob).GetClass() < j.(*Blob).GetClass()
}
*/