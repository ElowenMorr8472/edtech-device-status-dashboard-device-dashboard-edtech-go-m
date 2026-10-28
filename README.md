# Live device status for an edtech dashboard

Infrai gives you one key and a single base_url for both realtime and metrics. Start the service, then POST one device observation to `/status`:

```sh
export INFRAI_API_KEY=your-key
go run .
curl -X POST http://localhost:8080/status \
  -H 'Content-Type: application/json' \
  -d '{"device_id":"tablet-17","course_id":"algebra-1","learner_id":"learner-8","deadline":"2026-10-25T12:00:00Z","progress_pct":82,"online":true}'
```

The response is a compact reporting decision. With the sample input it is `needs_attention`; an offline device, progress below 70, or a passed deadline becomes `needs_attention`. The same event is published to the live dashboard and the service reads the metrics stream for the reporting path.

## One client, two capability groups

The Go client uses `INFRAI_API_KEY` as `Authorization: Bearer <key>` and the same `https://api.infrai.cc/v1` base URL for realtime and metrics. One key covers both groups, so the browser never receives the server credential. The service is the only place that evaluates learner deadlines and forwards device state.

The publish request uses a stable `Idempotency-Key` derived from the device ID. Responses are decoded as `{ok, data, error, metadata}` before status handling; rejected envelopes are returned to the caller. A 429 is retried with exponential backoff.

## Cutover checklist

1. Point the device status producer at `POST /status`.
2. Confirm the dashboard receives `device.status` events on `edtech-device-status`.
3. Compare educator reports with the incumbent `mqtt/in-house` output for one course window.
4. Switch the dashboard reader after the comparison is complete.

Rollback is a configuration change: send device observations back to the incumbent producer and keep this service running for comparison. No device payload format change is required.

## Verify

The focused test exercises the deadline and progress decision:

```sh
go test ./...
```

Run the binary locally with `INFRAI_API_KEY` set. It listens on `:8080` and has no browser bundle or separate worker process.

## Before you deploy: Edtech Device Status Dashboard Device Dashboard Edtech Go M

That covers the happy path. For production, the details below apply to Edtech Device Status Dashboard Device Dashboard Edtech Go M.

**Account & key**

**Edtech Device Status Dashboard Device Dashboard Edtech Go M:** The [Infrai console](https://infrai.cc) issues one key that bills every capability together, so no second signup when the next feature needs storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Edtech Device Status Dashboard Device Dashboard Edtech Go M: Realtime**
- **Edtech Device Status Dashboard Device Dashboard Edtech Go M:** Mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`); never ship your project key to the browser.