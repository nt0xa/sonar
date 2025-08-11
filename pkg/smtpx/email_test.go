package smtpx_test

import (
	"net/mail"
	"os"
	"testing"
	"time"

	"github.com/nt0xa/sonar/pkg/smtpx"
	"github.com/stretchr/testify/require"
)

func ptr[T any](v T) *T {
	return &v
}

func TestEmailParse(t *testing.T) {
	tests := []struct {
		file  string
		email smtpx.Email
	}{
		{
			"1.eml",
			smtpx.Email{},
		},
		{
			"2.eml",
			smtpx.Email{
				Subject: "test",
				To: []*mail.Address{
					{
						Name:    "Test Test",
						Address: "test@example.com",
					},
				},
				From: []*mail.Address{
					{
						Name:    "John Doe",
						Address: "john.doe@mail.com",
					},
				},
				Cc:   nil,
				Bcc:  nil,
				Date: ptr(time.Date(2025, 8, 5, 15, 48, 5, 0, time.UTC)),
				Text: "Test email",
			},
		},
		{
			"3.eml",
			smtpx.Email{
				Subject: "🚀 Your GitHub launch code",
				To: []*mail.Address{
					{
						Address: "user@test.com",
					},
				},
				From: []*mail.Address{
					{
						Name:    "GitHub",
						Address: "noreply@github.com",
					},
				},
				Cc:   nil,
				Bcc:  nil,
				Date: ptr(time.Date(2025, 3, 4, 10, 39, 22, 0, time.UTC)),
				Text: `Here's your GitHub launch code!

Continue signing up for GitHub by entering the code below:

123123123

You can enter it by visiting the link below:

You’re receiving this email because you recently created a new GitHub account. If this wasn’t you, please ignore this email.

Not able to enter the code? Paste the following link into your browser:

---
Sent with <3 by GitHub.
GitHub, Inc. 88 Colin P Kelly Jr Street
San Francisco, CA 94107`,
			},
		},
		{
			"4.eml",
			smtpx.Email{
				Subject: "🚀 Your GitHub launch code",
				To: []*mail.Address{
					{
						Address: "user@test.com",
					},
				},
				From: []*mail.Address{
					{
						Name:    "GitHub",
						Address: "noreply@github.com",
					},
				},
				Cc:   nil,
				Bcc:  nil,
				Date: ptr(time.Date(2025, 3, 4, 10, 39, 22, 0, time.UTC)),
				Text: `Here's your GitHub launch code!

Continue signing up for GitHub by entering the code below:

123123123

[Open GitHub](https://github.com/account_verifications?verification=&via_launch_code_email=true)

Not able to enter the code? Paste the following link into your browser: 

[https://github.com/account_verifications/confirm/](https://github.com/account_verifications/confirm/)

[Terms](https://docs.github.com/articles/github-terms-of-service/) ・
[Privacy](https://docs.github.com/articles/github-privacy-policy/) ・
[Sign in to GitHub](https://github.com/login) 

You’re receiving this email because you recently created a new GitHub account. If this wasn’t you, please ignore this email.

GitHub, Inc. ・88 Colin P Kelly Jr Street ・San Francisco, CA 94107`,
			},
		},
		{
			"5.eml",
			smtpx.Email{
				Subject: "Security alert",
				To: []*mail.Address{
					{
						Address: "user1.asdfgh@gmail.com",
					},
				},
				From: []*mail.Address{
					{
						Name:    "Google",
						Address: "no-reply@accounts.google.com",
					},
				},
				Cc:   nil,
				Bcc:  nil,
				Date: ptr(time.Date(2025, 8, 7, 14, 44, 11, 0, time.UTC)),
				Text: `[image: Google]
A new sign-in on Google Pixel 7


user1.asdfgh@gmail.com
We noticed a new sign-in to your Google Account on a Google Pixel 7 device.
If this was you, you don’t need to do anything. If not, we’ll help you
secure your account.
Check activity
<https://accounts.google.com/AccountChooser?Email=user1.asdfgh@gmail.com&continue=https://myaccount.google.com/alert/nt/1754577851000?rfn%3D325%26rfnc%3D1%26eid%3D7123123412341234123%26et%3D0>
You can also see security activity at
https://myaccount.google.com/notifications
You received this email to let you know about important changes to your
Google Account and services.
© 2025 Google LLC, 1600 Amphitheatre Parkway, Mountain View, CA 94043, USA`,
			},
		},
		{
			"6.eml",
			smtpx.Email{
				Subject: "Security alert",
				To: []*mail.Address{
					{
						Address: "user1.asdfgh@gmail.com",
					},
				},
				From: []*mail.Address{
					{
						Name:    "Google",
						Address: "no-reply@accounts.google.com",
					},
				},
				Cc:   nil,
				Bcc:  nil,
				Date: ptr(time.Date(2025, 8, 7, 14, 44, 11, 0, time.UTC)),
				Text: `A new sign-in on Google Pixel 7 
[user1.asdfgh@gmail.com] 
We noticed a new sign-in to your Google Account on a Google Pixel 7 device. If this was you, you don’t need to do anything. If not, we’ll help you secure your account.[Check activity](https://accounts.google.com/AccountChooser?Email=user1.asdfgh@gmail.com&continue=https://myaccount.google.com/alert/nt/1754577851000?rfn%3D325%26rfnc%3D1%26eid%3D7123123412341234123%26et%3D0)

You can also see security activity at
[https://myaccount.google.com/notifications]

You received this email to let you know about important changes to your Google Account and services.
© 2025 Google LLC, [1600 Amphitheatre Parkway, Mountain View, CA 94043, USA]`,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.file, func(t *testing.T) {
			b, err := os.ReadFile("test-data/" + test.file)
			require.NoError(t, err)

			email := smtpx.Parse(string(b))

			require.Equal(t, test.email.Subject, email.Subject)
			require.Equal(t, test.email.From, email.From)
			require.Equal(t, test.email.To, email.To)
			require.Equal(t, test.email.Date, email.Date)
			require.Equal(t, test.email.Text, email.Text)
		})
	}
}
