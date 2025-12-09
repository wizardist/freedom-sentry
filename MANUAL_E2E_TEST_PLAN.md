# Manual End-to-End Test Plan for Freedom-Sentry

**Version:** 1.0
**Date:** 2025-12-09
**Purpose:** Comprehensive manual testing guide for freedom-sentry suppression bot

---

## Table of Contents
1. [Prerequisites](#prerequisites)
2. [Test Environment Setup](#test-environment-setup)
3. [Test Scenarios](#test-scenarios)
4. [Verification Procedures](#verification-procedures)
5. [Expected Results](#expected-results)
6. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Required Access
- [ ] MediaWiki instance (non-production test wiki)
- [ ] Account with `suppressrevision` permission
- [ ] API access token with appropriate permissions
- [ ] Ability to create and edit pages on the test wiki

### Required Tools
- [ ] freedom-sentry binary (compiled from source or Docker image)
- [ ] curl or similar HTTP client (for verification)
- [ ] Access to MediaWiki Special:RecentChanges
- [ ] Access to MediaWiki API endpoint

### Knowledge Requirements
- Basic understanding of MediaWiki structure
- Familiarity with MediaWiki suppression/oversight features
- Command-line proficiency

---

## Test Environment Setup

### 1. Create Non-Production Suppression List

#### 1.1 Create the List Page
1. Log into your test MediaWiki instance
2. Navigate to a page for the suppression list (e.g., `User:TestBot/SuppressionList`)
3. Create the page with initial content:
   ```
   Test Page Alpha
   Test Page Beta
   ```
4. Save the page
5. **Verify:** Page exists and contains the two test page names

#### 1.2 Create Test Pages
1. Create the following test pages with initial content:
   - `Test Page Alpha` - "Initial content for Alpha"
   - `Test Page Beta` - "Initial content for Beta"
   - `Test Page Gamma` - "Initial content for Gamma" (intentionally NOT on list)
2. **Verify:** All three pages exist and are visible

### 2. Configure the Bot

#### 2.1 Set Environment Variables
```bash
export ACCESS_TOKEN="your-access-token-here"
export API_ENDPOINT="https://your-test-wiki.org/api.php"
export LIST_NAME="User:TestBot/SuppressionList"
export WIKI_DOMAIN="your-test-wiki.org"
export LOG_LEVEL="DEBUG"
export BATCHING_SUPPRESSOR_PERIOD="1s"
export BATCHING_SUPPRESSOR_SIZE="50"
```

#### 2.2 Verify Configuration
```bash
# Verify the list is accessible
curl -s "$API_ENDPOINT?action=query&titles=$LIST_NAME&prop=revisions&rvprop=content&format=json"
```
**Expected:** JSON response with page content showing "Test Page Alpha" and "Test Page Beta"

---

## Test Scenarios

### Scenario 1: Initial Startup and Full Scan

**Objective:** Verify bot performs initial full scan and suppresses existing revisions

#### Pre-Test Setup
1. Make edits to both listed pages:
   - Edit `Test Page Alpha`: Add line "Edit before bot startup"
   - Edit `Test Page Beta`: Add line "Edit before bot startup"
2. Note revision IDs from Special:RecentChanges
3. **Verify:** Both edits are currently visible (not suppressed)

#### Test Steps
1. Start the bot:
   ```bash
   ./freedom-sentry
   ```
2. Observe startup logs for:
   - Configuration validation
   - Token validation (checking suppressrevision rights)
   - Initial full scan execution
   - Connection to EventStreams

#### Verification
1. Wait 5 seconds for full scan to complete
2. Check Special:RecentChanges:
   - Edits to `Test Page Alpha` should show as suppressed
   - Edits to `Test Page Beta` should show as suppressed
   - Editor names should be hidden
   - Edit comments should be hidden
3. Check bot logs for:
   - "Starting initial full scan" message
   - Suppression API calls for both pages
   - No error messages

**Pass Criteria:**
- [ ] Bot starts successfully
- [ ] Initial full scan completes
- [ ] All revisions on listed pages are suppressed
- [ ] No errors in logs

---

### Scenario 2: Real-Time Suppression via EventStreams

**Objective:** Verify bot detects and suppresses edits in real-time

#### Test Steps
1. Ensure bot is running from Scenario 1
2. Make a new edit to `Test Page Alpha`: Add line "Real-time test edit"
3. Note the exact time of edit
4. Make another edit to `Test Page Beta`: Add line "Second real-time test"

#### Verification
1. Monitor bot logs for:
   - EventStreams event received
   - Suppression action triggered
2. Check Special:RecentChanges within 5 seconds:
   - Both new edits should be suppressed
   - Timestamp from edit to suppression should be < 2 seconds

#### Timing Test
1. Record time of edit (T1)
2. Continuously refresh Special:RecentChanges
3. Record time when suppression appears (T2)
4. Calculate latency: T2 - T1

**Pass Criteria:**
- [ ] Bot detects edits via EventStreams
- [ ] Edits are suppressed within 2 seconds
- [ ] Suppression is complete (user + comment hidden)
- [ ] Logs show batching suppressor processing

---

### Scenario 3: Negative Test - Unlisted Page

**Objective:** Verify bot does NOT suppress edits to pages not on the list

#### Test Steps
1. Ensure bot is running
2. Edit `Test Page Gamma`: Add line "This should NOT be suppressed"
3. Wait 10 seconds

#### Verification
1. Check Special:RecentChanges:
   - Edit to `Test Page Gamma` should be VISIBLE
   - Username should be visible
   - Edit comment should be visible
2. Check bot logs:
   - Should NOT show suppression action for Gamma
   - May show filtering/skipping logic

**Pass Criteria:**
- [ ] Edit to unlisted page remains visible
- [ ] Bot correctly filters out unlisted pages
- [ ] No suppression attempts in logs

---

### Scenario 4: Bot Edit Filtering

**Objective:** Verify bot ignores edits made by other bots

#### Pre-Test Setup
1. You need a second bot account or ability to mark edits as bot edits
2. If not available, skip this scenario

#### Test Steps
1. Make a bot edit to `Test Page Alpha` (use bot flag)
2. Observe bot behavior

#### Verification
1. Check bot logs:
   - Should show event filtering for bot edits
2. Verify the bot edit is NOT suppressed

**Pass Criteria:**
- [ ] Bot edits are filtered out
- [ ] No suppression attempts for bot edits

---

### Scenario 5: Dynamic List Updates

**Objective:** Verify bot responds to changes in the suppression list

#### Test Steps
1. Ensure bot is running
2. Edit the suppression list page (`User:TestBot/SuppressionList`)
3. Add a new line:
   ```
   Test Page Alpha
   Test Page Beta
   Test Page Gamma
   ```
4. Save the list
5. Wait 30 seconds for cache purge and list reload
6. Make an edit to `Test Page Gamma`: Add line "Now on the list"

#### Verification
1. Check bot logs for:
   - Detection of list page edit
   - Cache purge trigger
   - List reload
   - Full scan execution for new pages
2. Check Special:RecentChanges:
   - New edit to Gamma should be suppressed
3. Verify existing revisions on Gamma (before list update) are also suppressed

**Pass Criteria:**
- [ ] Bot detects list updates
- [ ] Cache is purged
- [ ] New pages are added to monitoring
- [ ] Subsequent edits to newly listed pages are suppressed
- [ ] Historical revisions are suppressed

---

### Scenario 6: Batching Behavior

**Objective:** Verify multiple edits are batched together efficiently

#### Test Steps
1. Ensure bot is running with `BATCHING_SUPPRESSOR_PERIOD=1s`
2. Quickly make 5 edits to `Test Page Alpha` within 1 second:
   - Edit 1: "Batch test 1"
   - Edit 2: "Batch test 2"
   - Edit 3: "Batch test 3"
   - Edit 4: "Batch test 4"
   - Edit 5: "Batch test 5"
3. Wait 3 seconds

#### Verification
1. Check bot logs:
   - Should show batching suppressor accumulating revisions
   - Should show a single API call with multiple revision IDs
   - Look for log message indicating batch size (e.g., "suppressing 5 revisions")
2. Verify all 5 edits are suppressed

**Pass Criteria:**
- [ ] Multiple edits are batched
- [ ] Single API call handles multiple revisions
- [ ] All edits are suppressed
- [ ] Batch flush occurs within configured period

---

### Scenario 7: Already Suppressed Revisions

**Objective:** Verify bot skips already-suppressed revisions

#### Pre-Test Setup
1. Manually suppress an edit to `Test Page Beta` using MediaWiki's native tools
2. Note the revision ID

#### Test Steps
1. Restart the bot to trigger initial full scan
2. Observe logs during startup

#### Verification
1. Check bot logs:
   - Should show filtering logic skipping already-suppressed revisions
   - Should NOT attempt to suppress the pre-suppressed revision again
2. Verify no duplicate suppression API calls

**Pass Criteria:**
- [ ] Bot detects already-suppressed revisions
- [ ] No redundant suppression attempts
- [ ] Filtering works correctly

---

### Scenario 8: Periodic Full Scan

**Objective:** Verify 15-minute periodic scan executes correctly

#### Test Steps
1. Start bot and note the timestamp
2. Let bot run for 16 minutes
3. Monitor logs

#### Verification
1. Check bot logs at ~15 minute mark:
   - Should show scheduled full scan execution
   - Should check all pages in the list
   - Should complete without errors
2. If any unsuppressed revisions exist, they should be caught and suppressed

**Pass Criteria:**
- [ ] Periodic scan executes every 15 minutes
- [ ] Scan covers all listed pages
- [ ] Scan completes successfully

---

### Scenario 9: Invalid/Non-Existent Pages

**Objective:** Verify graceful handling of invalid pages in the list

#### Test Steps
1. Edit the suppression list to include a non-existent page:
   ```
   Test Page Alpha
   Test Page Beta
   NonExistentPage123456
   ```
2. Wait for list reload (or restart bot)
3. Monitor logs

#### Verification
1. Check bot logs:
   - Should show API errors or warnings for non-existent page
   - Should NOT crash or stop processing other pages
   - Should continue monitoring other valid pages
2. Make an edit to `Test Page Alpha` - should still be suppressed

**Pass Criteria:**
- [ ] Bot handles non-existent pages gracefully
- [ ] Error logging is clear
- [ ] Bot continues operating normally
- [ ] Valid pages are still monitored

---

### Scenario 10: EventStreams Reconnection

**Objective:** Verify bot recovers from EventStreams connection loss

#### Test Steps
1. Start bot normally
2. Simulate network interruption (if possible):
   - Block access to EventStreams endpoint temporarily
   - Or use firewall rules
3. Restore connection after 30 seconds
4. Make an edit to `Test Page Alpha`

#### Verification
1. Check bot logs:
   - Should show EventStreams disconnection
   - Should show reconnection attempts with backoff
   - Should successfully reconnect
   - Should resume processing events
2. Verify the edit after reconnection is still suppressed

**Pass Criteria:**
- [ ] Bot detects connection loss
- [ ] Automatic reconnection is attempted
- [ ] Exponential backoff is applied
- [ ] Bot resumes normal operation after reconnection
- [ ] No events are missed (or caught by periodic scan)

---

## Verification Procedures

### Manual API Verification

You can verify suppression status directly via the MediaWiki API:

```bash
# Get revision details with suppression info
curl -s "$API_ENDPOINT?action=query&prop=revisions&revids=REVISION_ID&rvprop=ids|flags|timestamp|user|comment&format=json" | jq

# If suppressed, the response will show:
# - "userhidden": true
# - "commenthidden": true
# - "suppressed": true
```

### Log Analysis

Key log patterns to look for:

**Successful Operation:**
```
level=INFO msg="Starting freedom-sentry bot"
level=INFO msg="Token validation successful" rights="[suppressrevision, ...]"
level=INFO msg="Starting initial full scan"
level=DEBUG msg="Suppressing revisions" count=X
level=INFO msg="EventStreams client connected"
```

**Issues:**
```
level=ERROR msg="Failed to suppress revisions" error="..."
level=WARN msg="EventStreams connection lost" reason="..."
level=ERROR msg="API request failed" action="..." error="..."
```

---

## Expected Results

### Success Criteria Summary

All test scenarios should pass with:
- ✅ Real-time suppression latency < 2 seconds
- ✅ All edits to listed pages are suppressed
- ✅ Edits to unlisted pages are NOT suppressed
- ✅ Bot edits are filtered and not suppressed
- ✅ List updates are detected within 30 seconds
- ✅ Batching reduces API calls for rapid edits
- ✅ No duplicate suppressions
- ✅ Graceful error handling
- ✅ Automatic reconnection on failures
- ✅ Periodic scans execute every 15 minutes

### Performance Benchmarks

- **Startup time:** < 5 seconds
- **Suppression latency:** < 2 seconds (from edit to suppression)
- **Batching period:** ~1 second (configurable)
- **List cache TTL:** 24 hours
- **EventStreams reconnect:** Exponential backoff 1s → 60s

---

## Troubleshooting

### Bot Won't Start

**Symptom:** Bot crashes immediately on startup

**Checks:**
1. Verify `ACCESS_TOKEN` is set and valid
2. Verify `API_ENDPOINT` is accessible
3. Check token has `suppressrevision` permission:
   ```bash
   curl -s "$API_ENDPOINT?action=query&meta=userinfo&uiprop=rights&format=json" \
     -H "Authorization: Bearer $ACCESS_TOKEN" | jq '.query.userinfo.rights'
   ```
4. Check bot logs for specific error messages

### Edits Not Being Suppressed

**Symptom:** Edits remain visible after 10+ seconds

**Checks:**
1. Verify page is actually in the suppression list
2. Check bot logs for event processing
3. Verify EventStreams connection is active (look for "connected" log)
4. Check if edit is marked as bot edit (filtered out)
5. Verify suppression API calls are succeeding (no 403/401 errors)
6. Check network connectivity to wiki

### High Latency

**Symptom:** Suppression takes > 5 seconds

**Checks:**
1. Check EventStreams connection (may be reconnecting)
2. Verify `BATCHING_SUPPRESSOR_PERIOD` is not too high
3. Check MediaWiki API response times
4. Look for rate limiting or throttling in logs
5. Verify network latency between bot and wiki

### List Updates Not Detected

**Symptom:** New pages added to list are not monitored

**Checks:**
1. Verify 30 seconds have passed (cache purge delay)
2. Check bot logs for list page edit detection
3. Verify list page name matches `LIST_NAME` config
4. Try manual cache purge by restarting bot
5. Check list page format (one page title per line)

### False Positives

**Symptom:** Unlisted pages are being suppressed

**Checks:**
1. Verify suppression list content (may have unintended entries)
2. Check for whitespace or formatting issues in list
3. Review bot logs for page matching logic
4. Verify page titles match exactly (case-sensitive)

---

## Test Sign-Off

**Test Execution Date:** _______________

**Tester Name:** _______________

**Environment:** _______________

| Scenario | Pass | Fail | Notes |
|----------|------|------|-------|
| 1. Initial Startup | ☐ | ☐ | |
| 2. Real-Time Suppression | ☐ | ☐ | |
| 3. Unlisted Page | ☐ | ☐ | |
| 4. Bot Edit Filtering | ☐ | ☐ | |
| 5. Dynamic List Updates | ☐ | ☐ | |
| 6. Batching Behavior | ☐ | ☐ | |
| 7. Already Suppressed | ☐ | ☐ | |
| 8. Periodic Full Scan | ☐ | ☐ | |
| 9. Invalid Pages | ☐ | ☐ | |
| 10. Reconnection | ☐ | ☐ | |

**Overall Result:** ☐ PASS ☐ FAIL

**Comments:**
_______________________________________________
_______________________________________________
_______________________________________________

---

## Appendix: Quick Reference Commands

### Start Bot
```bash
./freedom-sentry
```

### Start Bot (Skip Initial Scan)
```bash
./freedom-sentry --skip-init-fullscan
```

### Enable Debug Logging
```bash
export LOG_LEVEL="DEBUG"
./freedom-sentry
```

### Check Suppression Status via API
```bash
curl -s "$API_ENDPOINT?action=query&prop=revisions&revids=$REV_ID&rvprop=ids|flags|user|comment&format=json" | jq
```

### Monitor Recent Changes
```bash
# Via browser
open "https://your-test-wiki.org/wiki/Special:RecentChanges"

# Via API
curl -s "$API_ENDPOINT?action=query&list=recentchanges&rcprop=title|ids|user|comment|flags&format=json" | jq
```

### Update Suppression List
1. Navigate to `https://your-test-wiki.org/wiki/$LIST_NAME`
2. Click "Edit"
3. Add/remove page titles (one per line)
4. Save
5. Wait 30 seconds for bot to detect change

---

**Document Version Control:**
- v1.0 (2025-12-09): Initial test plan creation
