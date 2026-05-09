package matchmaking

// FindMatch tries to find a compatible pair
func (q *Queue) FindMatch() (*QueueEntry, *QueueEntry) {
	for i := 0; i < len(q.entries); i++ {
		a := q.entries[i]

		if a.Status != StatusWaiting {
			continue
		}

		for j := i + 1; j < len(q.entries); j++ {
			b := q.entries[j]

			if b.Status != StatusWaiting {
				continue
			}

			// matching condition
			if a.Preference == b.Gender && b.Preference == a.Gender {
				return a, b
			}
		}
	}

	return nil, nil
}