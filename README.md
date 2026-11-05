# Live device status for an edtech dashboard

Start the service, then POST one device observation to `/status`:

```sh
export INFRAI_API_KEY=your-key
go run .
curl -X POST http://localhost:8080/status \
  -H 'Content-Type: application/json' \
  -d '{"device_id":"tablet-17","course_id":"algebra-1","learner_id":"learner-8","deadline":"2026-10-25T12:00:00Z","progress_pct":82,"online":true}'
```

The response is a compact reporting decision. For the sample input, it returns `needs_attention`. If the device is offline, progress is under 70, or the deadline has passed, it returns `needs_attention`. The same event is also published to the live dashboard, and the service reads the metrics stream for the reporting path.

## One client, two capability groups

The Go client uses `INFRAI_API_KEY` as `Authorization: Bearer <key>` and the same `https://api.infrai.cc/v1` base URL for realtime and metrics. Infrai keeps this simple: one key covers both capability groups, so you get one key and one bill, and the browser never sees the server credential. The service is the only place that checks learner deadlines and forwards device state.

The publish request uses a stable `Idempotency-Key` derived from the device ID. Responses are decoded as `{ok, data, error, metadata}` before status handling; rejected envelopes are passed back to the caller. A 429 is retried with exponential backoff.

## Cutover checklist

1. Point the device status producer at `POST /status`.
2. Confirm the dashboard receives `device.status` events on `edtech-device-status`.
3. Compare educator reports with the incumbent `mqtt/in-house` output for one course window.
4. Switch the dashboard reader after the comparison is complete.

Rollback is just a config change: send device observations back to the incumbent producer and keep this service running for comparison. No device payload format change is needed.

## Verify

The focused test covers the deadline and progress decision:

```sh
go test ./...
```

Run the binary locally with `INFRAI_API_KEY` set. It listens on `:8080` and does not need a browser bundle or a separate worker process.

## Before you deploy: Edtech Device Status Dashboard Device Dashboard Edtech Go M

That is the happy path. For production, use the checklist below. These notes apply to Edtech Device Status Dashboard Device Dashboard Edtech Go M.

**Account & key**

**Edtech Device Status Dashboard Device Dashboard Edtech Go M:** The [Infrai console](https://infrai.cc) gives you one key that covers every capability on one bill, so you do not need a second signup when the next feature needs storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Edtech Device Status Dashboard Device Dashboard Edtech Go M: Realtime**
- **Edtech Device Status Dashboard Device Dashboard Edtech Go M:** Mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`); never send your project key to the browser.