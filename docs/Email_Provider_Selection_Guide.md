# Email Provider Selection Guide

**Document Version:** 1.0
**Last Updated:** 2025-01-22
**For:** Gassigeher Application Administrators

---

## Overview

The Gassigeher application supports two email sending methods:
1. **Gmail API** (OAuth2) - Original method, using Google's API
2. **SMTP** (Username/Password) - Standard email protocol, works with any provider

This guide helps you choose the right email provider for your deployment.

---

## Quick Decision Matrix

| Your Situation | Recommended Provider |
|----------------|----------------------|
| Small shelter (<100 emails/day) | **Gmail API** |
| Using Strato email hosting | **SMTP (Strato)** |
| Using Office365 | **SMTP (Office365)** |
| Need >100 emails/day | **SMTP** |
| Already have Google Workspace | **Gmail API** |
| Prefer simple setup | **SMTP** |
| Want best deliverability | **Gmail API** |
| Corporate IT requirements | **SMTP (Office365)** |
| Custom email server | **SMTP (Custom)** |

---

## Detailed Comparison

### Gmail API

**✅ Advantages:**
- **Excellent deliverability** - Google's infrastructure
- **Free tier** - 100 emails/day (2,000 with Workspace)
- **No spam issues** - Trusted sender
- **Official API** - Well-maintained and documented
- **OAuth2 security** - No password storage

**❌ Disadvantages:**
- **Complex setup** - Requires Google Cloud Console
- **OAuth2 complexity** - Refresh tokens, API credentials
- **Daily limit** - 100 emails/day for free accounts
- **Google dependency** - Requires Google account

**Best For:**
- Small to medium shelters
- Existing Google users
- Maximum email deliverability
- Budget-conscious deployments

**Setup Time:** ~15-30 minutes (one-time)

---

### SMTP (Strato)

**✅ Advantages:**
- **Simple setup** - Just username and password
- **German provider** - Strato is popular in Germany
- **Higher limits** - Depends on your plan (typically 500-5000/day)
- **Use existing** - If you already have Strato email
- **Full control** - Your email infrastructure

**❌ Disadvantages:**
- **Requires Strato account** - Must have email package
- **Monthly cost** - Depends on Strato plan
- **SPF/DKIM setup** - May need DNS configuration
- **Deliverability** - Depends on configuration

**Best For:**
- German shelters
- Existing Strato customers
- Higher email volume
- German-language support

**Setup Time:** ~5-10 minutes

---

### SMTP (Office365)

**✅ Advantages:**
- **Enterprise-grade** - Microsoft infrastructure
- **High limits** - 10,000 emails/day
- **Corporate integration** - Part of Office365 suite
- **Reliability** - Microsoft SLA
- **Global presence** - Worldwide data centers

**❌ Disadvantages:**
- **Requires Office365** - Must have subscription
- **Monthly cost** - Part of Office365 pricing
- **Corporate focus** - May be overkill for small shelters
- **Authentication** - May require app password with 2FA

**Best For:**
- Large shelters
- Organizations already using Office365
- High email volume (>1,000/day)
- Corporate environments
- Enterprise support needs

**Setup Time:** ~5-10 minutes

---

### SMTP (Gmail SMTP)

**✅ Advantages:**
- **Simpler than Gmail API** - No OAuth2 complexity
- **Same Gmail infrastructure** - Good deliverability
- **No Google Cloud setup** - Just username and app password
- **Familiar** - Everyone knows Gmail

**❌ Disadvantages:**
- **Same limits as API** - 500/day (free), 2,000/day (Workspace)
- **Requires app password** - Must enable 2FA first
- **Less secure than API** - Password-based vs OAuth2
- **Google dependency** - Still requires Google account

**Best For:**
- Small shelters wanting Gmail without OAuth2 complexity
- Temporary setups
- Testing SMTP functionality
- Migrating from Gmail API

**Setup Time:** ~5 minutes

---

### SMTP (Custom Server)

**✅ Advantages:**
- **Complete control** - Your own infrastructure
- **No limits** - Configure as needed
- **Privacy** - Emails never leave your infrastructure
- **Customization** - Full configuration control

**❌ Disadvantages:**
- **Complex setup** - Server administration required
- **Maintenance** - Must maintain email server
- **Deliverability** - Must configure SPF/DKIM/DMARC
- **Spam risk** - Improper setup may result in spam classification

**Best For:**
- Organizations with existing email infrastructure
- On-premises deployments
- Specific compliance requirements
- Technical teams

**Setup Time:** Varies (depends on existing infrastructure)

---

## Feature Comparison Table

| Feature | Gmail API | SMTP (Strato) | SMTP (Office365) | SMTP (Gmail) | SMTP (Custom) |
|---------|-----------|---------------|------------------|--------------|---------------|
| **Daily Limit** | 100 (free) / 2,000 (Workspace) | 500-5,000 (plan dependent) | 10,000 | 500 (free) / 2,000 (Workspace) | Unlimited |
| **Setup Complexity** | ⭐⭐⭐ High | ⭐ Easy | ⭐ Easy | ⭐ Easy | ⭐⭐⭐ High |
| **Monthly Cost** | Free / $6/user | €5-15/month | €10.50/user | Free / $6/user | Server costs |
| **Authentication** | OAuth2 | Username/Password | Username/Password | App Password | Username/Password |
| **Deliverability** | ⭐⭐⭐⭐⭐ Excellent | ⭐⭐⭐⭐ Good | ⭐⭐⭐⭐⭐ Excellent | ⭐⭐⭐⭐⭐ Excellent | ⭐⭐⭐ Varies |
| **German Support** | ✅ Yes | ✅ Yes (Native) | ✅ Yes | ✅ Yes | ✅ Yes |
| **2FA Required** | No | No | Optional | Yes | No |
| **API vs SMTP** | API | SMTP | SMTP | SMTP | SMTP |
| **Port** | N/A | 465 (SSL) | 587 (TLS) | 587 (TLS) | Configurable |

---

## Cost Analysis

### Small Shelter (50 emails/day)
- **Gmail API:** **€0/month** ✅ Best choice
- **SMTP (Strato):** €5/month
- **SMTP (Office365):** €10.50/month
- **SMTP (Gmail):** €0/month (alternative)

### Medium Shelter (200 emails/day)
- **Gmail API:** €6/month (requires Workspace) ⚠️ At limit
- **SMTP (Strato):** **€10/month** ✅ Better value
- **SMTP (Office365):** €10.50/month
- **SMTP (Gmail):** €6/month (requires Workspace)

### Large Shelter (1,000 emails/day)
- **Gmail API:** €6/month (Workspace) ⚠️ May hit limits
- **SMTP (Strato):** €15/month ✅ Good option
- **SMTP (Office365):** **€10.50/month** ✅ Best value
- **SMTP (Gmail):** €6/month (Workspace) ⚠️ May hit limits

---

## Security Comparison

### Gmail API
- ✅ **Most Secure** - OAuth2, no password storage
- ✅ Refresh tokens (can be revoked)
- ✅ Scoped permissions
- ✅ Regular security updates from Google

### SMTP (All Providers)
- ⚠️ **Password-based** - Must protect credentials
- ✅ TLS/SSL encryption in transit
- ✅ Credentials in environment variables (not in code)
- ⚠️ Password compromise risk

**Recommendation:** Gmail API is more secure due to OAuth2, but SMTP is secure if you:
- Use strong passwords
- Enable 2FA on email account
- Use app passwords (not regular passwords)
- Keep credentials in environment variables
- Use TLS/SSL for all connections

---

## Deliverability Considerations

### Factors Affecting Deliverability

**Gmail API:**
- ✅ Excellent reputation (Google infrastructure)
- ✅ SPF/DKIM automatic
- ✅ Low spam risk

**SMTP (Strato/Office365):**
- ✅ Good reputation (established providers)
- ⚠️ SPF/DKIM may need configuration
- ⚠️ Medium spam risk if misconfigured

**SMTP (Custom):**
- ⚠️ Reputation depends on configuration
- ❌ Must manually configure SPF/DKIM/DMARC
- ❌ High spam risk if misconfigured

### DNS Configuration (SMTP Only)

For best deliverability with SMTP, configure:

**1. SPF Record:**
```
v=spf1 include:_spf.strato.de ~all
```

**2. DKIM:**
- Usually configured automatically by provider
- Check provider documentation

**3. DMARC:**
```
v=DMARC1; p=none; rua=mailto:admin@yourdomain.com
```

**Note:** Strato and Office365 handle most of this automatically.

---

## Migration Between Providers

### Gmail API → SMTP

**Difficulty:** Easy
**Time:** 5 minutes
**Steps:**
1. Get SMTP credentials from your email provider
2. Update `.env` file with SMTP settings
3. Change `EMAIL_PROVIDER=smtp`
4. Restart application

See: [Migration Guide](Email_Migration_Guide.md)

### SMTP → Gmail API

**Difficulty:** Medium
**Time:** 30 minutes
**Steps:**
1. Create Google Cloud project
2. Enable Gmail API
3. Create OAuth2 credentials
4. Generate refresh token
5. Update `.env` with Gmail API settings
6. Change `EMAIL_PROVIDER=gmail`
7. Restart application

See: [Gmail API Setup Guide](Gmail_API_Setup_Guide.md)

### Between SMTP Providers

**Difficulty:** Easy
**Time:** 5 minutes
**Steps:**
1. Update SMTP credentials in `.env`
2. Update `SMTP_HOST`, `SMTP_PORT`, `SMTP_USE_TLS`, `SMTP_USE_SSL`
3. Restart application

---

## Recommendations by Use Case

### Scenario: New Gassigeher Deployment

**If you:**
- Are setting up Gassigeher for the first time
- Send <100 emails/day
- Want minimal cost
- Can spend 30 minutes on setup

**Use:** **Gmail API**

**Why:** Best deliverability, free, and well-tested.

---

### Scenario: Existing Strato Customer

**If you:**
- Already have Strato email hosting
- Send 100-5,000 emails/day
- Want simple setup
- Prefer German support

**Use:** **SMTP (Strato)**

**Why:** Use your existing infrastructure, simple setup, higher limits.

---

### Scenario: Corporate/Enterprise

**If you:**
- Organization uses Office365
- Send >1,000 emails/day
- Need enterprise support
- Have IT department

**Use:** **SMTP (Office365)**

**Why:** Enterprise-grade, high limits, corporate integration.

---

### Scenario: Budget-Conscious

**If you:**
- Very limited budget
- Send <100 emails/day
- Okay with setup complexity
- Want long-term free option

**Use:** **Gmail API**

**Why:** Free tier sufficient, excellent deliverability, no ongoing costs.

---

### Scenario: High Volume

**If you:**
- Send >1,000 emails/day
- Need reliability
- Have budget for email service
- Want professional appearance

**Use:** **SMTP (Office365)** or **SMTP (Strato - Premium Plan)**

**Why:** High limits, reliability, professional infrastructure.

---

## Testing Your Choice

Before committing to a provider, test it:

1. **Register a test user** in Gassigeher
2. **Check email delivery**
   - Does email arrive?
   - Check spam folder
   - Verify formatting
3. **Test German umlauts** - ä, ö, ü, ß should display correctly
4. **Test HTML rendering** - Email should look professional
5. **Check delivery time** - Should arrive within 1-2 minutes
6. **Test all email types**:
   - Verification email
   - Welcome email
   - Booking confirmation
   - Password reset

---

## Getting Help

### Provider-Specific Support

**Gmail API:**
- Google Cloud Console: https://console.cloud.google.com
- Gmail API Docs: https://developers.google.com/gmail/api

**Strato:**
- Strato Support: https://www.strato.de/support
- Phone: +49 30 300 146 - 0

**Office365:**
- Microsoft 365 Admin Center: https://admin.microsoft.com
- Support: Built into admin portal

### Gassigeher Application Support

- GitHub Issues: https://github.com/your-org/gassigeher/issues
- Documentation: See [docs/](../docs/) folder

---

## Conclusion

**Best Overall:** Gmail API for small deployments, SMTP (Strato/Office365) for higher volume

**Easiest Setup:** SMTP (any provider)

**Most Secure:** Gmail API (OAuth2)

**Best Deliverability:** Gmail API or SMTP (Office365)

**Best Value:** Gmail API (free tier) or SMTP (Strato) for higher volume

**Choose based on:**
1. Your current email provider
2. Daily email volume
3. Budget
4. Technical expertise
5. Security requirements

---

**Last Updated:** 2025-01-22
**Version:** 1.0
**For Questions:** See application documentation or contact support
