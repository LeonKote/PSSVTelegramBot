package api

const (
	scale    = "scale=1280:720"
	codec    = "libx264"
	preset   = "ultrafast"
	flag     = "+faststart"
	pix      = "yuv420p"
	fVideo   = "mp4"
	metadata = "title=Recorded Video"

	frame   = 1        // только один кадр
	fImg    = "image2" // формат одиночного изображения
	quality = 1        // качество JPEG (1 — max, 31 — min)
	// formatImg = "mjpeg"  // формат изображения
)
