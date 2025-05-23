package handlers

// import (
// 	"backend/utils"
// 	"log"
// 	"net/http"
// )

// func DeleteMessage(w http.ResponseWriter, r *http.Request) {
// 	utils.EnableCORS(w)
// 	log.Println("🗑️：DeleteMessage")

// idStr := r.URL.Query().Get("id")
// log.Println(idStr)
// if idStr == "" {
// 	http.Error(w, "ID is required", http.StatusBadRequest)
// 	return
// }

// id, err := strconv.Atoi(idStr)
// if err != nil {
// 	http.Error(w, "Invalid ID", http.StatusBadRequest)
// 	return
// }

// if err := db.DB.Delete(&Message{}, id).Error; err != nil {
// 	http.Error(w, "DB delete error", http.StatusInternalServerError)
// 	return
// }

// w.WriteHeader(http.StatusOK)
// json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
// }
