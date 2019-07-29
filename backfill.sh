#!/usr/bin/env bash
#
# Backfills history in stellar-core DB.
# 
# Usage: ./backfill.sh <ledger_dest> <ledger_count>
#
# Example: 
#   ./backfill.sh 10000 1000 - will populate the DB with data from ledgers 9000 upto 10000

set -o errexit
set -o pipefail

ledger_dest=${1:-current}
ledger_count=${2:-max}
target_core_db=${3:-core}

echo "Starting backfill for ${ledger_count} ledgers upto ${ledger_dest}"

workdir="${ledger_dest}-${ledger_count}"

mkdir -p "${workdir}"
pushd ${workdir}

cat >stellar-core.cfg <<-ENDOFCONFIG
LOG_FILE_PATH="stellar-core.log"
BUCKET_DIR_PATH="buckets"
DATABASE="sqlite3://stellar.db"
NETWORK_PASSPHRASE="Public Global Stellar Network ; September 2015"
AUTOMATIC_MAINTENANCE_PERIOD=0
AUTOMATIC_MAINTENANCE_COUNT=0
[HISTORY.local]
get="cp history/{0} {1}"
[HISTORY.satoshipay_de]
get="curl -sf https://stellar-history-de-fra.satoshipay.io/{0} -o {1}"
[HISTORY.satoshipay_sg]
get="curl -sf https://stellar-history-sg-sin.satoshipay.io/{0} -o {1}"
[HISTORY.satoshipay_us]
get="curl -sf https://stellar-history-us-iowa.satoshipay.io/{0} -o {1}"
[HISTORY.sdf_1]
get="curl -sf http://history.stellar.org/prd/core-live/core_live_001/{0} -o {1}"
[HISTORY.sdf_2]
get="curl -sf http://history.stellar.org/prd/core-live/core_live_002/{0} -o {1}"
[HISTORY.sdf_3]
get="curl -sf http://history.stellar.org/prd/core-live/core_live_003/{0} -o {1}"
[QUORUM_SET]
THRESHOLD_PERCENT=66
VALIDATORS=[
  "GCGB2S2KGYARPVIA37HYZXVRM2YZUEXA6S33ZU5BUDC6THSB62LZSTYH sdf_1",
  "GCM6QMP3DLRPTAZW2UZPCPX2LF3SXWXKPMP3GKFZBDSF3QZGV2G5QSTK sdf_2",
  "GABMKJM6I25XI4K7U6XWMULOUQIQ27BCTMLS6BYYSOWKTBUXVRJSXHYQ sdf_3",
  "GAOO3LWBC4XF6VWRP5ESJ6IBHAISVJMSBTALHOQM2EZG7Q477UWA6L7U eno"
]
ENDOFCONFIG

ln -sf /var/lib/stellar/history history

stellar-core new-db
stellar-core catchup "${ledger_dest}/${ledger_count}"


# Temp workaround for constraint violation on copy
sqlite3 stellar.db 'delete from ledgerheaders where ledgerseq=1'

for table in ledgerheaders txhistory txfeehistory upgradehistory scphistory; do
  sqlite3 -header -csv stellar.db "select * from ${table}" | psql -c "\copy ${table} from stdin csv header;" "${target_core_db}"
done

popd