package mp

import (
	"testing"
)

var benchConfig *Config
var benchDetector *Detector

func init() {
	var err error
	benchConfig, err = loadConfigFromFile("../config.json")
	if err != nil {
		panic(err)
	}
	benchDetector = NewDetector(benchConfig)
}

// Бенчмарк для нормальных ников (должны проходить).
func BenchmarkCheckNormalNick(b *testing.B) {
	nicks := []string{
		"норм", "себастьян", "облако на небе", "илья", "магомед",
		"глебик", "давид", "хлеб", "облако", "надпись",
		"себастьянчик", "мебель", "оскорблять",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchDetector.Check(nicks[i%len(nicks)])
	}
}

// Бенчмарк для плохих ников (должны блокироваться).
func BenchmarkCheckBadNick(b *testing.B) {
	nicks := []string{
		"suka",
		"pidaras",
		"ебанутыми",
		"обоссыш",
		"хуепезд",
		"блядь",
		"ПИСЮН ДЛЯ ЖОПЫ",
		"ДРОЧКА",
		"3бaнyт",
		"3б@нyт",
		"п1зда",
		"х_у_й",
		"хуесосами",
		"x_y_e_c",
		"п---и---з---д---а",
		"хлебал твой рот",
		"я ебал твой рот",
		"Кот ебучий. Очень* ",
		"мать твоя шлюха",
		"пизда",
		"jopa",
		"пидрила",
		"я где-то далеко в ебенях",
		"zhe shi megahui",
		"уe64н",
		"россияхуйня",
		"пидоры себастьянчик",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, nick := range nicks {
			benchDetector.Check(nick)
		}
	}
}

// Бенчмарк для граничных случаев (короткие ники).
func BenchmarkCheckShortNick(b *testing.B) {
	nicks := []string{
		"a",
		"ab",
		"abc",
		"123",
		"",
		" ",
		"x",
		"y",
		"z",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, nick := range nicks {
			benchDetector.Check(nick)
		}
	}
}

// Бенчмарк для ников с лишними символами.
func BenchmarkCheckObfuscatedNick(b *testing.B) {
	nicks := []string{
		"п-ид0ры с-еб4__с тьян чик",
		"x_y_e_c",
		"3б@нyт",
		"п---и---з---д---а",
		"хлеб%ал твой рот",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, nick := range nicks {
			benchDetector.Check(nick)
		}
	}
}

// Бенчмарк для длинных ников.
func BenchmarkCheckLongNick(b *testing.B) {
	longNick := "этооченьдлинныйниккоторыйдолженпроверитьсянадлительностьиневызватьпроблемспроизводительностьюисодержитразныеслованапримерсебастьянкакойтохлеб"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchDetector.Check(longNick)
	}
}

// Бенчмарк для смешанных ников (латиница + кириллица).
func BenchmarkCheckMixedNick(b *testing.B) {
	nicks := []string{
		"russianРусский",
		"cEBASTYANсебастьян",
		"XLEBхлеб",
		"пидорPIDOR",
		"cykaBLYAT",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, nick := range nicks {
			benchDetector.Check(nick)
		}
	}
}

// Параллельный бенчмарк.
func BenchmarkCheckParallel(b *testing.B) {
	nick := "подозрительныйниккоторыйпроверяется"
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			benchDetector.Check(nick)
		}
	})
}