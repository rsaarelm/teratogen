// doc.go
//
// Copyright (C) 2013 Risto Saarelma
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

/*
Teratogen is a survival horror roguelike game.

Write the general documentation here.

- The map has hexagon geometry, but looks isometric due to tricks with the
tiles.

- The map geometry is noneuclidean. Map cells may be portals that lead to
arbitrary places elsewhere in the game world.

- The overall architecture is based on building the full application out of
stateful system objects that refer to each other in an acyclic graph. Each
system object generally has a single instance, and takes care of a single
aspect of the game, like rendering the world map or running field-of-view
computations.

- Animations in the game screen are triggered by game events, but do not
affect game events back or block new game events. You should be able to make
the game run as fast as you can hit the keys.

- The game assets are embedded into the game executable by packing them into a
.zip file and catenating the .zip file into the executable.

*/
package main
