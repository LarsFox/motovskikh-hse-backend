package main

import (
	"fmt"
	"motovskikh-hse-backend/mp"
)

func main() {
	config := mp.DefaultConfig()
	detector := mp.NewDetector(config)

	// Проверка ников.
	nicks := []string{
		"еbик",
		"сцука",
		"suka",
		"pidaras",
		"ебанутыми",
		"обоссыш",
		"Это хуес",
		"хуепезд",
		"блядь",
		"норм",
		"ПИСЮН ДЛЯ ЖОПЫ",
		"ДРОЧКА",
		"себастьян",
		"облако на небе",
		"илья",
		"магомед",
		"глебик",
		"давид",
		"хлеб",
		"3бaнyт",
		"ебоаик",
		"3б@нyт",
		"п1зда",
		"х_у_й",
		"сцууука",
		"хуесосами",
		"x_y_e_c",
		"п---и---з---д---а",
		"ебанутейшими",
		"хyйй",
		"пiздa",
		"себастьян",
		"облако",
		"ридор",
		"хлебал твой рот",
		"я ебал твой рот",
		"надпись",
		"Кот ебучий. Очень* ",
	  "хлеб%ал твой рот",
	  "мать твоя шлюха",
	  "пизда",
		"jopa",
		"",
		"я где-то далеко в ебенях",
	}
	
	fmt.Println("Проверка ников\n")
	for _, nick := range nicks {
		result := detector.Check(nick)
		if result.IsAllowed {
			fmt.Printf("%s - разрешен\n", nick)
		} else {
			fmt.Printf("%s - ЗАПРЕЩЕН (причина: %s", nick, result.Reason)
			if result.Word != "" {
				fmt.Printf(", слово: %s", result.Word)
			}
			fmt.Println(")")
		}
	}
}