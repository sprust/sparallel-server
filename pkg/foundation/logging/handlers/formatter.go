package handlers

import "log/slog"

func makeMessageByRecord(r slog.Record) string {
	return r.Time.Format("2006-01-02 15:04:05.000") + " " + r.Level.String() + " " + r.Message
}
