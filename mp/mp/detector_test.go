package mp

import (
	"testing"
)

func TestDetector(t *testing.T) {
	config, err := loadConfigFromFile("../config.json")
	if err != nil {
		t.Fatalf("Error")
	}

	detector := NewDetector(config)

	tests := []struct {
		nick     string
		expected bool // true - разрешен, false - запрещен.
	} {
		// Запрещенные.
		{"suka", false},
		{"pidàras", false},
		{"ебанутыми", false},
		{"обоссыш", false},
		{"Это хуес", false},
		{"хуепезд", false},
		{"блядь", false},
		{"ПИСЮН ДЛЯ ЖОПЫ", false},
		{"ДРОЧКА", false},
		{"3бaнyт", false},
		{"3б@нyт", false},
		{"п1зда", false},
		{"х_у_й", false},
		{"хуесосами", false},
		{"x_y_e_c", false},
		{"п---и---з---д---а", false},
		{"ебанутейшими", false},
		{"хyйй", false},
		{"пiздa", false},
		{"хлебал твой рот", false},
		{"я ебал твой рот", false},
		{"Кот ебучий. Очень* ", false},
		{"хлеб%ал твой рот", false},
		{"мать твоя шлюха", false},
		{"пизда", false},
		{"jopa", false},
		{"пидрила", false},
		{"я где-то далеко в ебенях", false},
		{"zhe shi megahui", false},
		{"уe64н", false},
		{"россияхуйня", false},
		{"пидоры хлебатить", false},
		{"п-ид0ры с-еб4__с тьян чик", false},


		// Разрешенные.
		{"норм", true},
		{"себастьян", true},
		{"облако на небе", true},
		{"илья", true},
		{"магомед", true},
		{"глебик", true},
		{"давид", true},
		{"хлеб", true},
		{"облако", true},
		{"надпись", true},
		{"себастьянчик", true},
		{"мебель", true},
		{"оскорблять", true},
	}

	for _, tt := range tests {
		t.Run(tt.nick, func(t *testing.T) {
			result := detector.Check(tt.nick)
			if result.IsAllowed != tt.expected {
				t.Errorf("nick: %q\n  ожидалось: %v\n", tt.nick, tt.expected)
			}
		})
	}
}