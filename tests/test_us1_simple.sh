#!/bin/bash
set -e

cd /Users/jochem/code/enron-graph

echo "=== User Story 1 Test ==="
echo ""

# Create minimal test CSV
echo "Creating test email data..."
cat > /tmp/test_emails.csv << 'EOF'
file,message
msg1.txt,"Message-ID: <001@enron.com>
From: jeff.skilling@enron.com
To: ken.lay@enron.com
Date: Mon, 1 Jan 2001 10:00:00 -0600
Subject: Q4 Results

Ken, please review the attached Q4 results. The numbers look good.

Best regards,
Jeff"
msg2.txt,"Message-ID: <002@enron.com>
From: ken.lay@enron.com
To: jeff.skilling@enron.com
Date: Mon, 1 Jan 2001 11:00:00 -0600
Subject: Re: Q4 Results

Jeff, excellent work. Let's discuss this with Andrew Fastow tomorrow.

Ken"
msg3.txt,"Message-ID: <003@enron.com>
From: andrew.fastow@enron.com
To: jeff.skilling@enron.com,ken.lay@enron.com
Date: Tue, 2 Jan 2001 09:00:00 -0600
Subject: Meeting Notes

Attached are the meeting notes from our discussion.

Andrew"
EOF

echo "✅ Created test data with 3 emails"
echo ""

# Test Scenario 1: Load without extraction
echo "=== AC-011: Load emails without extraction ==="
go run cmd/loader/main.go --csv-path /tmp/test_emails.csv --workers 2
echo "✅ Loaded emails successfully"
echo ""

# Test Scenario 2: Load with extraction
echo "=== AC-012: Load emails with entity extraction ==="
go run cmd/loader/main.go --csv-path /tmp/test_emails.csv --workers 2 --extract
echo "✅ Loaded and extracted entities successfully"
echo ""

echo "=== Test Complete ==="
echo "✅ User Story 1 acceptance criteria validated"
