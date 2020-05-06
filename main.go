package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	fail2banCurrentFailed = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fail2ban_current_failed",
			Help: "Number of failed attempts",
		},
		[]string{"jail"},
	)

	fail2banTotalFailed = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fail2ban_total_failed",
			Help: "Total Number of failed attempts",
		},
		[]string{"jail"},
	)

	fail2banCurrentBanned = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fail2ban_current_banned",
			Help: "Number of current banned IPs",
		},
		[]string{"jail"},
	)

	fail2banTotalBanned = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fail2ban_total_banned",
			Help: "Total Number of banned IPs",
		},
		[]string{"jail"},
	)

	reg = prometheus.NewRegistry()
)

func initLog() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})
	log.Info("logger initialised")
}

func init() {
	initLog()
	reg.MustRegister(fail2banCurrentFailed)
	reg.MustRegister(fail2banCurrentBanned)
	reg.MustRegister(fail2banTotalFailed)
	reg.MustRegister(fail2banTotalBanned)
}

type jail struct {
	name          string
	currentfailed float64
	totalfailed   float64
	currentbanned float64
	totalbanned   float64
}

func processJailStat(fail2banstats string, regex string) float64 {
	re := regexp.MustCompile(regex)
	line := re.FindString(fail2banstats)
	statstring := strings.Split(line, ":")
	statvalue := strings.TrimSpace(statstring[1])
	statfloat, _ := strconv.ParseFloat(statvalue, 64)
	return statfloat
}

func jailProcess(jailname string) (*jail, error) {
	out, err := exec.Command("fail2ban-client", "status", jailname).Output()
	if err != nil {
		log.Error("Error finding jail:", jailname)
		return nil, err
	}
	fail2banstats := string(out)
	j := jail{name: jailname}
	j.currentfailed = processJailStat(fail2banstats, `.*Currently failed:.*`)
	j.totalfailed = processJailStat(fail2banstats, `.*Total failed:.*`)
	j.currentbanned = processJailStat(fail2banstats, `.*Currently banned:.*`)
	j.totalbanned = processJailStat(fail2banstats, `.*Total banned:.*`)
	return &j, nil
}

func jailList() string {
	out, err := exec.Command("fail2ban-client", "status").Output()
	if err != nil {
		fmt.Println(err)
	}
	re := regexp.MustCompile(`.*Jail list:.*`)
	jailsout := re.FindString(string(out))
	jailslist := strings.Split(jailsout, "`- Jail list:	")
	return jailslist[1]
}

func generateJailsArray() ([]jail, error) {
	var jails []jail
	for _, i := range strings.Split(jailList(), ",") {
		jailname := strings.Replace(i, " ", "", -1)
		j, err := jailProcess(string(jailname))
		if err != nil {
			log.Error("Cannot generate jails array")
			return nil, err
		}
		jails = append(jails, *j)
	}
	return jails, nil
}

func jailsHander(w http.ResponseWriter, r *http.Request) {

	jails, err := generateJailsArray()

	if err != nil {
		log.Error("Cannot generate jails array")
	} else {
		for _, jail := range jails {
			fail2banCurrentFailed.WithLabelValues(jail.name).Set(jail.currentfailed)
			fail2banCurrentBanned.WithLabelValues(jail.name).Set(jail.currentbanned)
			fail2banTotalFailed.WithLabelValues(jail.name).Set(jail.totalfailed)
			fail2banTotalBanned.WithLabelValues(jail.name).Set(jail.totalbanned)
		}
		promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(w, r)
	}
}

func jailHander(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	j, err := jailProcess(vars["jail"])
	if err != nil {
		log.Error("Cannot generate jails array")
		return
	}

	smallreg := prometheus.NewRegistry()
	smallreg.MustRegister(fail2banCurrentFailed)
	smallreg.MustRegister(fail2banCurrentBanned)
	smallreg.MustRegister(fail2banTotalFailed)
	smallreg.MustRegister(fail2banTotalBanned)

	fail2banCurrentFailed.WithLabelValues(j.name).Set(j.currentfailed)
	fail2banCurrentBanned.WithLabelValues(j.name).Set(j.currentbanned)
	fail2banTotalFailed.WithLabelValues(j.name).Set(j.totalfailed)
	fail2banTotalBanned.WithLabelValues(j.name).Set(j.totalbanned)

	promhttp.HandlerFor(smallreg, promhttp.HandlerOpts{}).ServeHTTP(w, r)
}

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/metrics", jailsHander)
	r.HandleFunc("/probe/{jail}", jailHander)
	log.Fatal(http.ListenAndServe(":8089", r))
}
