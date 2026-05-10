package base

import "time"

// Fresh небезопасная функция, обязательно оберни ее мьютексом!
func (r *Room) Fresh() {
	r.last = time.Now()
}

// GG небезопасная функция, обязательно оберни ее мьютексом!
func (r *Room) GG() {
	totalRows := make([]*totalScore, 0, len(r.Players))
	for playerID, player := range r.Players {
		player.Score = r.ScoreFunc(playerID)
		player.ready = false

		if player.IsSpectator() {
			continue
		}

		totalRows = append(totalRows, &totalScore{
			nick:     player.Nick,
			score:    r.FmtScoreFunc(player.Score),
			scoreVal: player.Score,
		})
	}

	r.totalScore = totalRows
	r.totalTime = time.Since(r.startedAt)
	r.State = RoomStateFinished
}
