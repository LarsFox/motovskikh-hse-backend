package api

import (
	"net/http"
)

func (m *Manager) hndlrCreateTestData(w http.ResponseWriter, r *http.Request) {
	if err := m.manager.CreateTestData(); err != nil {
		m.sendError(w, http.StatusInternalServerError, "Failed to create test data: "+err.Error())
		return
	}
	
	testIDs := []string{"1", "2", "europe", "asia"}
	availableTests := []map[string]interface{}{}
	
	for _, testID := range testIDs {
		test, err := m.manager.GetTest(testID)
		if err == nil && test != nil {
			availableTests = append(availableTests, map[string]interface{}{
				"id":   testID,
				"name": test.Name,
			})
		}
	}
	
	m.send(w, map[string]interface{}{
		"test_data_created": true,
		"message":           "Тестовые данные созданы",
		"available_tests":   availableTests,
	})
}