# E2E Test Status

Last full audit: _Not yet run_

## Summary

| Spec File                 | Passed | Failed | Total |
| ------------------------- | ------ | ------ | ----- |
| game-flow.spec.ts         | ?      | ?      | ?     |
| interactive-board.spec.ts | ?      | ?      | ?     |
| setup-phase.spec.ts       | ?      | ?      | ?     |
| robber.spec.ts            | ?      | ?      | ?     |
| trading.spec.ts           | ?      | ?      | ?     |
| development-cards.spec.ts | ?      | ?      | ?     |
| longest-road.spec.ts      | ?      | ?      | ?     |
| ports.spec.ts             | ?      | ?      | ?     |
| victory.spec.ts           | ?      | ?      | ?     |
| **TOTAL**                 | ?      | ?      | ?     |

---

## Failing Tests by Spec

### game-flow.spec.ts

_Update after audit_

### interactive-board.spec.ts

_Update after audit_

### setup-phase.spec.ts

_Update after audit_

### robber.spec.ts

_Update after audit_

### trading.spec.ts

_Update after audit_

### development-cards.spec.ts

_Update after audit_

### longest-road.spec.ts

_Update after audit_

### ports.spec.ts

_Update after audit_

### victory.spec.ts

_Update after audit_

---

## Root Cause Patterns

Track common failure patterns here:

- [ ] Missing backend command handler
- [ ] Missing frontend UI component
- [ ] WebSocket message not sent/received
- [ ] Selector `data-cy` attribute missing
- [ ] Race condition / timing issue
- [ ] State not updating correctly

---

## Fix Log

| Date | Test Fixed | Root Cause | Iteration |
| ---- | ---------- | ---------- | --------- |
|      |            |            |           |

---

## Notes

- Test artifacts (screenshots, traces) are in `frontend/test-results/`
- Run single spec: `cd frontend && npx playwright test <spec>.spec.ts`
- Run all with report: `cd frontend && npx playwright test --reporter=html`
