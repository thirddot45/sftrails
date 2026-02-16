#!/usr/bin/env bash
# Push mock votes to sftrails to simulate real activity.
# Usage: ./scripts/mock_votes.sh [base_url]

BASE="${1:-http://localhost:8080}"

vote() {
    local trail_id="$1" vote_type="$2" ip="$3" fp="$4"
    local code
    code=$(curl -s -o /dev/null -w "%{http_code}" \
        -X POST "$BASE/vote" \
        -H "X-Forwarded-For: $ip" \
        -d "trail_id=$trail_id&vote=$vote_type&fingerprint=$fp")
    echo "  trail=$trail_id vote=$vote_type ip=$ip -> $code"
}

echo "=== Pushing mock votes to $BASE ==="
echo ""

echo "Markham Park (1) - 5 rideable, 1 closed"
vote 1 open   10.0.1.1 fp-aaa1
vote 1 open   10.0.1.2 fp-aaa2
vote 1 open   10.0.1.3 fp-aaa3
vote 1 open   10.0.1.4 fp-aaa4
vote 1 open   10.0.1.5 fp-aaa5
vote 1 closed 10.0.1.6 fp-aaa6
echo ""

echo "Oleta River (2) - 1 rideable, 4 closed"
vote 2 closed 10.0.2.1 fp-bbb1
vote 2 closed 10.0.2.2 fp-bbb2
vote 2 closed 10.0.2.3 fp-bbb3
vote 2 closed 10.0.2.4 fp-bbb4
vote 2 open   10.0.2.5 fp-bbb5
echo ""

echo "Virginia Key (3) - 3 rideable, 3 closed (tied)"
vote 3 open   10.0.3.1 fp-ccc1
vote 3 open   10.0.3.2 fp-ccc2
vote 3 open   10.0.3.3 fp-ccc3
vote 3 closed 10.0.3.4 fp-ccc4
vote 3 closed 10.0.3.5 fp-ccc5
vote 3 closed 10.0.3.6 fp-ccc6
echo ""

echo "Quiet Waters (4) - 3 rideable"
vote 4 open 10.0.4.1 fp-ddd1
vote 4 open 10.0.4.2 fp-ddd2
vote 4 open 10.0.4.3 fp-ddd3
echo ""

echo "Halpatiokee (5) - 3 closed"
vote 5 closed 10.0.5.1 fp-eee1
vote 5 closed 10.0.5.2 fp-eee2
vote 5 closed 10.0.5.3 fp-eee3
echo ""

echo "Amelia Earhart (6) - 4 rideable, 2 closed"
vote 6 open   10.0.6.1 fp-fff1
vote 6 open   10.0.6.2 fp-fff2
vote 6 open   10.0.6.3 fp-fff3
vote 6 open   10.0.6.4 fp-fff4
vote 6 closed 10.0.6.5 fp-fff5
vote 6 closed 10.0.6.6 fp-fff6
echo ""

echo "Jonathan Dickinson (7) - no votes"
echo ""

echo "Riverbend (8) - 2 rideable, 1 closed"
vote 8 open   10.0.8.1 fp-hhh1
vote 8 open   10.0.8.2 fp-hhh2
vote 8 closed 10.0.8.3 fp-hhh3
echo ""

echo "Dyer Park (9) - 1 rideable, 2 closed"
vote 9 open   10.0.9.1 fp-iii1
vote 9 closed 10.0.9.2 fp-iii2
vote 9 closed 10.0.9.3 fp-iii3
echo ""

echo "Pinehurst (10) - 1 rideable"
vote 10 open 10.0.10.1 fp-jjj1
echo ""

echo "West Delray (11) - 1 closed"
vote 11 closed 10.0.11.1 fp-kkk1
echo ""

echo "Tree Tops (12) - 3 rideable, 1 closed"
vote 12 open   10.0.12.1 fp-lll1
vote 12 open   10.0.12.2 fp-lll2
vote 12 open   10.0.12.3 fp-lll3
vote 12 closed 10.0.12.4 fp-lll4
echo ""

echo "=== Done! Refresh your browser ==="
