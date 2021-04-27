set +eax

DATASET_DIRECTORY=dataset
if [ -d "$DATASET_DIRECTORY" ]
then
  echo "$DATASET_DIRECTORY directory exists and is not going to be created again."
else
  echo "$DATASET_DIRECTORY directory does not exist. Creating directory as necessary."
  mkdir -p $DATASET_DIRECTORY
fi

TRANCO_DATASET_URL="https://tranco-list.eu/download/WXW9/1000000"
TRANCO_DATASET_FILE="tranco-1m.csv"
if [ -e "$DATASET_DIRECTORY/$TRANCO_DATASET_FILE" ]
then
  echo "$TRANCO_DATASET_FILE exists and will not be downloaded again"
else
  echo "$TRANCO_DATASET_FILE does not exist and needs to be downloaded."
  read -p "Do you want to download the file now (y|Y|n|N) : " response
  if [[ $response == 'Y' || $response == 'y' ]]
  then
    echo "Fetching the Tranco dataset from $TRANCO_DATASET_URL"
    CURL_OK=true
    WGET_OK=true
    if ! command -v curl &> /dev/null
    then
        echo "curl could not be found. Proceeding to check for wget"
        CURL_OK=false
    fi
    if ! command -v curl &> /dev/null
    then
        echo "wget could not be found. Aborting. Please install curl/wget"
        WGET_OK=false
    fi
    echo "Command Exists [curl] $CURL_OK"
    echo "Command Exists [wget] $WGET_OK"
    if [[ $CURL_OK == true || $WGET_OK == true ]]
    then
      if [[ $CURL_OK == true ]]
      then
        curl $TRANCO_DATASET_URL --output "$DATASET_DIRECTORY/$TRANCO_DATASET_FILE"
      else
        wget -O "$DATASET_DIRECTORY/$TRANCO_DATASET_FILE" $TRANCO_DATASET_URL
      fi
      echo "File Download Complete and saved at $DATASET_DIRECTORY/$TRANCO_DATASET_FILE"
      awk -F, '{printf "%s\n", $2}' $DATASET_DIRECTORY/$TRANCO_DATASET_FILE > $DATASET_DIRECTORY/$TRANCO_DATASET_FILE.awk
      # Those terrible CR LF
      sed $'s/\r//' $DATASET_DIRECTORY/$TRANCO_DATASET_FILE.awk > $DATASET_DIRECTORY/$TRANCO_DATASET_FILE.sed
      mv $DATASET_DIRECTORY/$TRANCO_DATASET_FILE.sed $DATASET_DIRECTORY/$TRANCO_DATASET_FILE
      rm $DATASET_DIRECTORY/$TRANCO_DATASET_FILE.awk
    fi
  else
    echo "Aborting and not fetching the Tranco dataset"
  fi
fi
