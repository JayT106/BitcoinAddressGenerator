# This is an example how to generate the MultiSigP2SHAddress
# You can modify the input arguments with the json object format

curl -v -X POST "http://localhost:8080/v1/genMultiSigP2SHAddress" \
  -H "Content-type: application/json" \
  -H "Accept: application/json" \
  -d '{"n":"2","m":"3","publicKeys":"04a882d414e478039cd5b52a92ffb13dd5e6bd4515497439dffd691a0f12af9575fa349b5694ed3155b136f09e63975a1700c9f4d4df849323dac06cf3bd6458cd,046ce31db9bdd543e72fe3039a1f1c047dab87037c36a669ff90e28da1848f640de68c2fe913d363a51154a0c62d7adea1b822d05035077418267b1a1379790187,0411ffd36c70776538d079fbae117dc38effafb33304af83ce4894589747aee1ef992f63280567f52f5ba870678b4ab4ff6c8ea600bd217870a8b4f1f09f3a8e83"}'
