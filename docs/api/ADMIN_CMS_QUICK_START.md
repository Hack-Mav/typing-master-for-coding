# Admin CMS Quick Start Guide

## Accessing the Admin CMS

### Prerequisites
1. User account with `admin` or `moderator` role
2. Valid JWT authentication token
3. Access to the application at the configured URL

### Login
1. Navigate to the application login page
2. Enter admin credentials
3. Upon successful authentication, navigate to `/admin/dashboard`

## Admin Dashboard

**URL:** `/admin/dashboard`

### Overview Statistics
- **Total Users:** Current user count
- **Total Admins:** Admin and moderator count
- **Sessions (24h):** Recent activity
- **Total Content:** Combined count of languages, lessons, and snippets

### Quick Actions
- **Manage Languages** → `/admin/languages`
- **Manage Lessons** → `/admin/lessons`
- **Manage Snippets** → `/admin/snippets`
- **A/B Tests** → `/admin/ab-tests`

### Recent Activity
View the last 10 practice sessions with details:
- Session ID
- Mode (Zen, Timed Drill, etc.)
- Language
- Start time
- User ID

## Content Management

### 1. Lesson Builder

**URL:** `/admin/lessons/new` (create) or `/admin/lessons/:id/edit` (edit)

#### Creating a Lesson

1. **Basic Information:**
   - **Title:** Descriptive lesson name
   - **Language:** Select from available languages
   - **Difficulty:** 1 (Beginner) to 5 (Expert)
   - **Estimated Time:** 5-300 minutes

2. **Learning Objectives:**
   - Click "Add Objective" to add new objectives
   - Enter clear, measurable learning goals
   - Remove objectives with the "Remove" button

3. **Prerequisites:**
   - Enter comma-separated lesson IDs
   - Leave empty if no prerequisites

4. **Token Coverage Checklist:**
   - View language-specific tokens in the sidebar
   - Check tokens covered in this lesson
   - Progress bar shows completion percentage
   - Tokens are color-coded by difficulty:
     - **Green:** Level 1-2 (Beginner/Easy)
     - **Yellow:** Level 3 (Intermediate)
     - **Red:** Level 4-5 (Advanced/Expert)

5. **Save:**
   - Click "Save Lesson" to create/update
   - Automatic version creation on update

#### Token Coverage Best Practices
- Aim for 70%+ coverage for comprehensive lessons
- Start with basic tokens for beginner lessons
- Progress to advanced tokens in later lessons
- Ensure prerequisite lessons cover foundational tokens

### 2. YAML Validator

**URL:** `/admin/yaml-validator`

#### Validating YAML Content

1. **Enter YAML:**
   - Type or paste YAML content in the editor
   - Use "Load Sample" for a template

2. **Format:**
   - Click "Format YAML" to auto-format

3. **Validate:**
   - Click "Validate YAML"
   - View results in the right panel:
     - ✅ **Valid:** Green indicator, no errors
     - ❌ **Invalid:** Red indicator, error list

4. **Create Snippet:**
   - Fill in snippet details:
     - Title
     - Language (YAML, Python, JavaScript)
     - Difficulty (1-5)
     - Tags (comma-separated)
   - Click "Create Snippet"
   - Only enabled after successful validation

#### Validation Rules
- YAML structure must be valid
- Content length: 10-1500 characters
- Required fields: title, source code
- Accessibility tags auto-generated

### 3. Snippet Curation

**URL:** `/admin/snippets`

#### Filtering Snippets

1. **Search:**
   - Search by title, content, or tags
   - Real-time filtering

2. **Language Filter:**
   - Select specific language or "All Languages"

3. **Difficulty Filter:**
   - Filter by difficulty level (1-5)

4. **Tag Filter:**
   - Select from existing tags

#### Managing Snippets

**Individual Actions:**
- **Edit:** Click "Edit" button on snippet card
- **Delete:** Click "Delete" button (confirmation required)

**Bulk Actions:**
1. Select multiple snippets using checkboxes
2. Choose action:
   - **Add Tag:** Select tag from dropdown
   - **Delete Selected:** Remove multiple snippets

#### Snippet Details Display
Each snippet card shows:
- Title and language
- Difficulty badge (color-coded)
- Source code preview (truncated at 200 chars)
- Tags
- Estimated time
- Creation date

### 4. A/B Testing Framework

**URL:** `/admin/ab-tests`

#### Creating an A/B Test

1. **Test Configuration:**
   - **Name:** Descriptive test name
   - **Description:** What you're testing
   - **Test Type:**
     - `scoring_weights`: Test different scoring formulas
     - `ui_variant`: Test UI changes

2. **Test Parameters:**
   - **Duration:** Test length in days (1-30)
   - **Rollout Percentage:** User participation rate (1-100%)

3. **Variants:**
   - Default: Control + Variant A
   - Click "Add Variant" for more variants
   - For scoring_weights tests, configure:
     - **TWPM Weight:** 0-1
     - **Raw Accuracy Weight:** 0-1
     - **Syntax Accuracy Weight:** 0-1
     - **Backspace Penalty:** 0-1
     - **Idle Time Penalty:** 0-1

4. **Create:**
   - Click "Create Test"
   - Test starts immediately if status is "active"

#### Managing Tests

**Status Changes:**
- **Active:** Test is running
- **Paused:** Temporarily stopped
- **Completed:** Test finished

**View Results:**
- Click "View Results" on any test
- See participant count and conversion rates
- Compare variant performance

**Delete Test:**
- Click "Delete" (confirmation required)
- Removes test and all data

#### A/B Testing Best Practices
- Run tests for at least 7 days
- Use 10-20% rollout for initial tests
- Ensure variants are significantly different
- Monitor results regularly
- Complete tests before starting new ones

## Content Versioning

**URL:** `/admin/content/versioning`

### Viewing Version History

1. **Select Content:**
   - Click on any content item in the left panel
   - View version history in the right panel

2. **Version Details:**
   - Version number
   - Created by (user)
   - Creation timestamp
   - Change notes
   - Active status indicator (green dot)

3. **View Content:**
   - Click "View Content" to expand JSON
   - Review complete content snapshot

### Restoring a Version

1. Select content item
2. Find desired version in history
3. Click "Restore" button
4. Confirm restoration
5. New version created with restored content

### Version Management Best Practices
- Always add meaningful change notes
- Review version before restoring
- Test restored content immediately
- Keep version history for audit trail

## Content Validation & QA

**URL:** `/admin/content/validate`

### Running Validations

1. **Select Content Type:**
   - Languages
   - Lessons
   - Snippets

2. **Click Validate:**
   - "Validate All Languages"
   - "Validate All Lessons"
   - "Validate All Snippets"

3. **Review Results:**
   - Success: Green indicator
   - Failure: Red indicator with error list

### Validation Rules

**Languages:**
- ✓ Parser ID must be valid
- ✓ Grammar configuration must be valid JSON
- ✓ Whitespace rules must define valid boundaries
- ✓ Language name must be unique

**Lessons:**
- ✓ Title must be descriptive
- ✓ Language ID must reference existing language
- ✓ Token coverage must include basic syntax
- ✓ Estimated time: 5-300 minutes
- ✓ Prerequisites must reference existing lessons

**Snippets:**
- ✓ Title must be unique
- ✓ Source code must be valid for language
- ✓ YAML snippets must have valid syntax
- ✓ Content length: 10-1500 characters
- ✓ Checksum must match content
- ✓ Accessibility tags auto-generated

### Quality Assurance Guidelines
- All content must be reviewed before publication
- Code snippets should demonstrate best practices
- Lessons should have clear learning objectives
- Accessibility considerations must be included
- Version control must be maintained for all changes
- Regular validation checks should be performed

## Content Analytics

**URL:** `/admin/content/analytics`

### Overview Metrics

**Dashboard Cards:**
- **Total Sessions:** All practice sessions
- **Active Users:** Current user count
- **Avg Session Time:** Minutes per session
- **Retention Rate:** User retention percentage

### Popular Content

**Popular Languages:**
- Ranked by session count
- Shows usage trends

**Most Viewed Content:**
- Top lessons and snippets
- View counts
- Content type indicators

### Performance Details

**Content Performance Table:**
- **Views:** Total view count
- **Completions:** Successful completions
- **Success Rate:** Completion percentage
  - Green: ≥80%
  - Yellow: 60-79%
  - Red: <60%
- **Avg Time:** Average completion time
- **Trend:** Usage trend indicator
  - ↗ Increasing (green)
  - → Stable (gray)
  - ↘ Decreasing (red)
- **Last Accessed:** Most recent usage

### Engagement Metrics

**User Activity:**
- **DAU:** Daily Active Users
- **WAU:** Weekly Active Users
- **MAU:** Monthly Active Users
- **Retention Rate:** User retention percentage

### Optimization Recommendations

**Automated Insights:**
- 📈 **Content to Promote:** High engagement, needs visibility
- ⚠️ **Content to Review:** Declining usage or low success rate
- 🎯 **Improvement Opportunities:** Stable but underperforming
- ✅ **High Performers:** Consistently successful content

### Time Range Filtering
- Last 24 hours
- Last 7 days
- Last 30 days
- Last 90 days

## API Endpoints Reference

### Authentication
All admin endpoints require:
- Valid JWT token delivered as an HttpOnly, Secure cookie (`credentials: 'include'`)
- Admin or moderator role
- The backend routes are under `/api/v1/` and are protected by role middleware; there is no `/api/v1/admin/` prefix

### Dashboard
```
GET /api/v1/dashboard
```

### Languages
```
GET    /api/v1/languages
POST   /api/v1/languages
PUT    /api/v1/languages/:id
DELETE /api/v1/languages/:id
```

### Lessons
```
GET    /api/v1/lessons
POST   /api/v1/lessons
PUT    /api/v1/lessons/:id
DELETE /api/v1/lessons/:id
```

### Snippets
```
GET    /api/v1/snippets
POST   /api/v1/snippets
PUT    /api/v1/snippets/:id
DELETE /api/v1/snippets/:id
```

### Content Versioning
```
GET  /api/v1/content/versions/:contentType/:contentId
POST /api/v1/content/versions/:contentType/:contentId/restore/:version
```

### Content Validation
```
POST /api/v1/content/validate
```

### A/B Testing
```
GET    /api/v1/ab-tests
POST   /api/v1/ab-tests
PUT    /api/v1/ab-tests/:id
DELETE /api/v1/ab-tests/:id
GET    /api/v1/ab-tests/:id/results
```

## Troubleshooting

### Common Issues

**1. "Unauthorized" Error**
- Verify JWT token is valid
- Check user role is admin or moderator
- Re-login if token expired

**2. Validation Failures**
- Review validation error messages
- Check required fields are filled
- Verify data types match schema
- Ensure references (language IDs, etc.) exist

**3. Version Restore Issues**
- Confirm version exists
- Check user permissions
- Verify content integrity

**4. A/B Test Not Running**
- Check test status is "active"
- Verify rollout percentage > 0
- Ensure end date is in future

### Getting Help

**Error Messages:**
- Read error messages carefully
- Check browser console for details
- Review network requests in DevTools

**Data Issues:**
- Validate content before saving
- Check version history for changes
- Use validation tools to identify problems

**Performance:**
- Use time range filters to reduce data load
- Clear browser cache if UI is slow
- Check network connection

## Best Practices

### Content Creation
1. **Plan First:** Outline lesson structure before creating
2. **Validate Early:** Use validation tools during creation
3. **Test Content:** Practice with created content
4. **Iterate:** Use analytics to improve content
5. **Version Control:** Always add change notes

### Content Management
1. **Regular Audits:** Review content quarterly
2. **Monitor Analytics:** Check performance weekly
3. **Update Content:** Refresh based on usage data
4. **Quality Control:** Validate all content regularly
5. **Backup:** Version history provides automatic backup

### A/B Testing
1. **One Test at a Time:** Avoid overlapping tests
2. **Sufficient Duration:** Run for at least 7 days
3. **Clear Hypothesis:** Know what you're testing
4. **Document Results:** Record findings
5. **Implement Winners:** Apply successful variants

### Security
1. **Protect Credentials:** Never share admin passwords; use passwords with 8+ characters, mixed case, a digit, and a special character
2. **Regular Audits:** Review admin actions
3. **Least Privilege:** Grant admin access sparingly
4. **Monitor Activity:** Check recent activity regularly
5. **Logout:** Always logout when finished
6. **CSRF Protection:** Admin POST/PUT/DELETE requests must include the `X-CSRF-Token` header matching the `csrf_token` cookie
7. **Rate Limiting:** API requests are rate-limited per IP; Redis-backed when `REDIS_URL` is configured, otherwise in-memory
8. **Account Lockout:** Repeated failed login/MFA attempts (5 within 15 minutes) lock the account
9. **Email Verification:** Admin account emails must be verified before login is permitted
10. **Token Rotation:** Refresh tokens are single-use; a successful refresh invalidates the previous refresh token

## Support

For technical issues or questions:
1. Check this documentation
2. Review implementation summary: `ADMIN_CMS_IMPLEMENTATION_SUMMARY.md`
3. Check API documentation
4. Contact development team

---

**Last Updated:** December 2024
**Version:** 1.0
**Status:** Production Ready
