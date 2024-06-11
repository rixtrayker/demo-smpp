package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    Requests = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "app_requests_total",
            Help: "Total number of requests",
        },
        []string{"path"},
    )

    RequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "app_request_duration_seconds",
            Help:    "Histogram of request durations",
            Buckets: prometheus.DefBuckets,
        },
        []string{"path"},
    )

    ErrorRate = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "app_errors_total",
            Help: "Total number of errors",
        },
        []string{"path"},
    )

    ActiveSessions = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "app_active_sessions",
            Help: "Number of active sessions",
        },
    )

    SessionDuration = prometheus.NewHistogram(
        prometheus.HistogramOpts{
            Name:    "app_session_duration_seconds",
            Help:    "Histogram of session durations",
            Buckets: prometheus.DefBuckets,
        },
    )

    SentMessages = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "app_sent_messages_total",
            Help: "Total number of sent messages",
        },
        []string{"status", "gateway", "ported_or_new"},
    )

    PortedMessages = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "app_ported_messages_total",
            Help: "Total number of ported messages",
        },
        []string{"status", "from", "to"},
    )

    TotallyFailedPortedMessages = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "app_totally_failed_ported_messages_total",
            Help: "Number of totally failed ported messages",
        },
        []string{"gateway1", "gateway2", "gateway3"},
    )
    SubmitSMRespDuration = prometheus.NewHistogram(
        prometheus.HistogramOpts{
            Name:    "app_submit_sm_resp_duration_seconds",
            Help:    "Histogram of SubmitSMResp message durations",
            Buckets: prometheus.DefBuckets,
        },
    )
)

func init() {
    prometheus.MustRegister(Requests, RequestDuration, ErrorRate, ActiveSessions, SessionDuration, SentMessages, PortedMessages, TotallyFailedPortedMessages)
}

func StartPrometheusServer() {
    http.Handle("/metrics", promhttp.Handler())
    go func() {
        http.ListenAndServe(":2112", nil)
    }()
}
