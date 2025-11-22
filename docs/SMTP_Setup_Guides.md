# SMTP Setup Guides

**Document Version:** 1.0
**Last Updated:** 2025-01-22
**For:** Gassigeher Application Administrators

This document provides step-by-step setup instructions for popular SMTP email providers.

---

## Table of Contents

1. [Strato SMTP Setup](#1-strato-smtp-setup)
2. [Office365 SMTP Setup](#2-office365-smtp-setup)
3. [Gmail SMTP Setup](#3-gmail-smtp-setup)
4. [Generic SMTP Setup](#4-generic-smtp-setup)

---

## 1. Strato SMTP Setup

### Prerequisites
- Active Strato email package
- Email address created (e.g., `noreply@yourdomain.com`)
- Access to Strato customer portal

### Step 1: Get SMTP Credentials

1. Log in to Strato customer portal: https://www.strato.de/apps/CustomerService
2. Navigate to **Email** → **Email Addresses**
3. Select your email address
4. Note the SMTP settings:
   - **Host:** `smtp.strato.de`
   - **Port:** `465` (SSL) or `587` (TLS)
   - **Username:** Your full email address
   - **Password:** Your email password

### Step 2: Configure Gassigeher

Edit your `.env` file:

```bash
# Email Provider
EMAIL_PROVIDER=smtp

# Strato SMTP Configuration
SMTP_HOST=smtp.strato.de
SMTP_PORT=465
SMTP_USERNAME=noreply@yourdomain.com
SMTP_PASSWORD=your-email-password
SMTP_FROM_EMAIL=noreply@yourdomain.com
SMTP_USE_SSL=true
SMTP_USE_TLS=false

# Optional: BCC Admin Copy
EMAIL_BCC_ADMIN=admin@yourdomain.com
```

**Port 465 (Recommended for Strato):**
- Use `SMTP_USE_SSL=true`
- Use `SMTP_USE_TLS=false`

**Port 587 (Alternative):**
- Use `SMTP_USE_SSL=false`
- Use `SMTP_USE_TLS=true`

### Step 3: Restart Application

```bash
sudo systemctl restart gassigeher
```

### Step 4: Test Email Sending

1. Register a new test user
2. Check email delivery
3. Verify German umlauts display correctly
4. Check spam folder if email doesn't arrive

### Troubleshooting

**Email not arriving:**
- Check credentials are correct
- Verify email address is active in Strato
- Check spam folder
- Check application logs: `sudo journalctl -u gassigeher -n 50`

**Authentication failed:**
- Ensure username is full email address
- Verify password is correct
- Check if account is locked due to failed attempts

**Connection refused:**
- Verify port 465 or 587 is not blocked by firewall
- Check if ISP blocks SMTP ports
- Try alternative port (587 if using 465)

### Strato-Specific Notes

- SPF/DKIM configured automatically by Strato
- Daily sending limits depend on your package
- Support: +49 30 300 146 - 0

---

## 2. Office365 SMTP Setup

### Prerequisites
- Active Office365/Microsoft 365 subscription
- Email address configured in Office365
- Admin access or app password capability

### Step 1: Enable SMTP AUTH (If Not Enabled)

1. Log in to Microsoft 365 Admin Center: https://admin.microsoft.com
2. Navigate to **Settings** → **Org settings** → **Modern authentication**
3. Ensure **Authenticated SMTP** is enabled
4. Save changes

### Step 2: Create App Password (If Using 2FA)

1. Go to https://account.microsoft.com/security
2. Navigate to **Security** → **Advanced security options**
3. Select **App passwords**
4. Click **Create a new app password**
5. Copy the generated password (you can't see it again)

### Step 3: Get SMTP Settings

Office365 SMTP settings are standard:
- **Host:** `smtp.office365.com`
- **Port:** `587` (TLS)
- **Username:** Your full Office365 email address
- **Password:** Your account password or app password

### Step 4: Configure Gassigeher

Edit your `.env` file:

```bash
# Email Provider
EMAIL_PROVIDER=smtp

# Office365 SMTP Configuration
SMTP_HOST=smtp.office365.com
SMTP_PORT=587
SMTP_USERNAME=noreply@yourdomain.com
SMTP_PASSWORD=your-password-or-app-password
SMTP_FROM_EMAIL=noreply@yourdomain.com
SMTP_USE_TLS=true
SMTP_USE_SSL=false

# Optional: BCC Admin Copy
EMAIL_BCC_ADMIN=admin@yourdomain.com
```

**Important:**
- Always use port 587 with TLS for Office365
- Use app password if 2FA is enabled
- Must use `SMTP_USE_TLS=true`

### Step 5: Restart Application

```bash
sudo systemctl restart gassigeher
```

### Step 6: Test Email Sending

1. Register a new test user
2. Check email delivery
3. Verify all email types work
4. Check formatting and German characters

### Troubleshooting

**Authentication failed:**
- Use app password if 2FA is enabled
- Verify username is full email address (user@domain.com)
- Check password doesn't contain special characters that need escaping
- Ensure account is licensed (not disabled)

**SMTP AUTH disabled error:**
- Enable authenticated SMTP in admin center
- Wait 15 minutes for changes to propagate
- Contact Microsoft 365 admin if you're not an admin

**Connection timeout:**
- Verify port 587 is not blocked
- Check firewall rules
- Ensure server can reach smtp.office365.com

**Sending limits exceeded:**
- Office365 limit: 10,000 emails/day
- Rate limit: 30 messages/minute
- Monitor usage in admin center

### Office365-Specific Notes

- Sending limit: 10,000 emails/day per mailbox
- Excellent deliverability (Microsoft infrastructure)
- SPF/DKIM configured automatically
- Support: Through Microsoft 365 admin portal

---

## 3. Gmail SMTP Setup

### Prerequisites
- Gmail account (free or Workspace)
- 2-Factor Authentication enabled
- App password generated

### Step 1: Enable 2-Factor Authentication

1. Go to https://myaccount.google.com/security
2. Under **Signing in to Google**, select **2-Step Verification**
3. Follow the setup process
4. Verify 2FA is enabled

### Step 2: Generate App Password

1. Go to https://myaccount.google.com/apppasswords
2. Select app: **Mail**
3. Select device: **Other (Custom name)**
4. Enter name: **Gassigeher Application**
5. Click **Generate**
6. **Copy the 16-character password** (spaces will be removed)
7. Save it securely (you won't see it again)

### Step 3: Get SMTP Settings

Gmail SMTP settings:
- **Host:** `smtp.gmail.com`
- **Port:** `587` (TLS)
- **Username:** Your Gmail address
- **Password:** The 16-character app password (not your regular password)

### Step 4: Configure Gassigeher

Edit your `.env` file:

```bash
# Email Provider
EMAIL_PROVIDER=smtp

# Gmail SMTP Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-16-char-app-password
SMTP_FROM_EMAIL=your-email@gmail.com
SMTP_USE_TLS=true
SMTP_USE_SSL=false

# Optional: BCC Admin Copy
EMAIL_BCC_ADMIN=admin@yourdomain.com
```

**Important:**
- Must use app password, NOT regular Gmail password
- Remove spaces from app password
- Always use port 587 with TLS

### Step 5: Restart Application

```bash
sudo systemctl restart gassigeher
```

### Step 6: Test Email Sending

1. Register a new test user
2. Check email delivery
3. Monitor sending limits

### Troubleshooting

**Authentication failed:**
- Ensure using app password, not regular password
- Verify 2FA is enabled
- Check app password is copied correctly (no spaces)
- Try generating new app password

**"Less secure apps" error:**
- This error means you're not using app password
- Regular password authentication is disabled
- Must use app password

**Daily limit exceeded:**
- Free Gmail: 500 emails/day
- Google Workspace: 2,000 emails/day
- Error: "You have reached a limit for sending mail"
- Wait 24 hours for reset

**Connection refused:**
- Check port 587 is not blocked
- Verify firewall allows SMTP
- Some ISPs block port 587

### Gmail-Specific Notes

- Free account: 500 emails/day limit
- Workspace account: 2,000 emails/day limit
- Excellent deliverability
- SPF/DKIM configured automatically
- Alternative to Gmail API (simpler setup, same infrastructure)

---

## 4. Generic SMTP Setup

Use these instructions for any SMTP server not covered above (custom servers, other providers, etc.).

### Step 1: Gather SMTP Information

Contact your email provider or check their documentation for:
- SMTP host/server address
- SMTP port
- Authentication method (usually username/password)
- TLS/SSL requirements
- Your email address and password

### Step 2: Determine Port and Encryption

Common configurations:

**Port 587 (STARTTLS - Recommended):**
```bash
SMTP_PORT=587
SMTP_USE_TLS=true
SMTP_USE_SSL=false
```

**Port 465 (SSL/TLS):**
```bash
SMTP_PORT=465
SMTP_USE_SSL=true
SMTP_USE_TLS=false
```

**Port 25 (Insecure - Not Recommended):**
```bash
SMTP_PORT=25
SMTP_USE_TLS=false
SMTP_USE_SSL=false
```

### Step 3: Configure Gassigeher

Edit your `.env` file:

```bash
# Email Provider
EMAIL_PROVIDER=smtp

# Generic SMTP Configuration
SMTP_HOST=mail.yourdomain.com
SMTP_PORT=587  # or 465, 25
SMTP_USERNAME=your-email@yourdomain.com
SMTP_PASSWORD=your-password
SMTP_FROM_EMAIL=noreply@yourdomain.com
SMTP_USE_TLS=true  # Adjust based on port
SMTP_USE_SSL=false # Adjust based on port

# Optional: BCC Admin Copy
EMAIL_BCC_ADMIN=admin@yourdomain.com
```

### Step 4: Test Connection

Before restarting the application, you can test SMTP connection using:

```bash
# Test SMTP connection (Linux/Mac)
telnet smtp.yourdomain.com 587

# Or using openssl for TLS
openssl s_client -connect smtp.yourdomain.com:587 -starttls smtp
```

### Step 5: Restart Application

```bash
sudo systemctl restart gassigeher
```

### Step 6: Monitor Logs

```bash
# Watch logs for errors
sudo journalctl -u gassigeher -f

# Check last 50 lines
sudo journalctl -u gassigeher -n 50
```

### Common Generic SMTP Issues

**Connection refused:**
- Verify host and port are correct
- Check firewall rules
- Ensure SMTP server is accessible from your network

**Authentication failed:**
- Verify username (may be full email or just username part)
- Check password is correct
- Some servers require specific authentication methods

**TLS/SSL handshake failed:**
- Verify TLS/SSL settings match server requirements
- Port 587 usually needs `SMTP_USE_TLS=true`
- Port 465 usually needs `SMTP_USE_SSL=true`
- Check if server certificate is valid

**Emails going to spam:**
- Configure SPF record for your domain
- Enable DKIM signing (check provider docs)
- Set up DMARC policy
- Ensure reverse DNS is configured

### DNS Configuration for Better Deliverability

**SPF Record (TXT):**
```
v=spf1 a mx ip4:YOUR_SERVER_IP ~all
```

**DMARC Record (TXT at _dmarc.yourdomain.com):**
```
v=DMARC1; p=none; rua=mailto:admin@yourdomain.com
```

**DKIM:**
- Contact your email provider for DKIM setup
- Usually involves adding a TXT record with public key

---

## Testing Checklist

After configuring any SMTP provider, test these scenarios:

- [ ] **User Registration** - Verification email arrives
- [ ] **Email Verification** - Link works, email confirmed
- [ ] **Password Reset** - Reset email arrives, link works
- [ ] **Booking Confirmation** - Booking email arrives
- [ ] **Booking Cancellation** - Cancellation email arrives
- [ ] **German Characters** - ä, ö, ü, ß display correctly
- [ ] **HTML Formatting** - Email looks professional, not broken
- [ ] **Delivery Time** - Emails arrive within 1-2 minutes
- [ ] **Spam Check** - Emails not going to spam folder
- [ ] **BCC Admin Copy** - Admin receives copy if configured

---

## Security Best Practices

### All SMTP Providers

1. **Use Strong Passwords**
   - At least 16 characters
   - Mix of uppercase, lowercase, numbers, symbols
   - Never reuse passwords

2. **Enable 2FA**
   - Always enable 2-Factor Authentication on email account
   - Use app passwords instead of regular passwords

3. **Secure Credential Storage**
   - Store credentials in environment variables
   - Never commit `.env` file to git
   - Use secrets management in production

4. **Use TLS/SSL**
   - Always use encrypted connections
   - Port 587 (STARTTLS) or 465 (SSL)
   - Never use port 25 unencrypted

5. **Monitor Usage**
   - Check email logs regularly
   - Watch for unusual sending patterns
   - Set up alerts for authentication failures

6. **Rotate Credentials**
   - Change passwords periodically
   - Regenerate app passwords annually
   - Update all instances when rotating

---

## Getting Help

### Provider-Specific Support

**Strato:** +49 30 300 146 - 0 | https://www.strato.de/support

**Office365:** Microsoft 365 admin center | https://admin.microsoft.com

**Gmail:** Google Workspace support | https://workspace.google.com/support

### Application Support

**Gassigeher:**
- Documentation: `/docs` folder
- GitHub Issues: [Your GitHub repository]
- Email: [Your support email]

---

## Conclusion

Choose the SMTP provider that best fits your needs:
- **Strato**: Best for German shelters with existing Strato hosting
- **Office365**: Best for enterprise, high volume
- **Gmail SMTP**: Best for small shelters wanting Gmail without OAuth2 complexity
- **Generic**: For custom servers and other providers

All providers work identically with Gassigeher - just choose based on your requirements, budget, and technical expertise.

---

**Last Updated:** 2025-01-22
**Version:** 1.0
**Next:** See [Email Provider Selection Guide](Email_Provider_Selection_Guide.md) for choosing a provider
