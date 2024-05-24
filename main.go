package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)
type apiConfig struct {
	fileserverHits int
	mu             sync.Mutex 
}
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		cfg.mu.Lock()
		cfg.fileserverHits++;
		cfg.mu.Unlock()
		next.ServeHTTP(w,r);
	})
}
func readinessHandler(w http.ResponseWriter, r *http.Request){
	w.Header().Add("Content-Type","text/plain; charset=utf-8");
	w.WriteHeader(http.StatusOK);
	w.Write([]byte(http.StatusText(http.StatusOK)));
}
func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request){
	w.Header().Add("Content-Type","text/plain; charset=utf-8");
	w.WriteHeader(http.StatusOK);
	w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileserverHits)))
}
func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
    cfg.mu.Lock()
    cfg.fileserverHits = 0
    cfg.mu.Unlock()
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Hits reset to 0"))
}
func main(){
	fmt.Println("Code begins!");
	apiCfg := apiConfig{
		fileserverHits: 0,
	}
	mux := http.NewServeMux();
	fileServer := http.FileServer(http.Dir("."));
	mux.Handle("/app/*",apiCfg.middlewareMetricsInc(http.StripPrefix("/app",fileServer)));
	mux.HandleFunc("/api/healthz", readinessHandler);
	mux.HandleFunc("/api/metrics", apiCfg.metricsHandler);
	mux.HandleFunc("/api/reset", apiCfg.resetHandler)
	srv := &http.Server{
		Addr: ":8080",
		Handler: mux,
	}
	log.Println("Starting server on PORT: 8080");
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Couldn't start the server %s\n",err)
	}

}