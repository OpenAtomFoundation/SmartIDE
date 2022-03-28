
echo 'ide.sh............start'

if [ -d "./opensumi-release/" ];then
 sudo rm -rf opensumi-release
else
  echo 'opensumi-release............不存在'
fi

if [ -d "./opensumi-extension/" ];then
 sudo rm -rf opensumi-extension
else
  echo 'opensumi-extension...........不存在'
fi

sudo mkdir opensumi-release opensumi-extension
sudo chmod -R 777 opensumi-release
sudo chmod -R 777 opensumi-extension


# 解压opensumi-release
sudo tar -zxf opensumi-release.tar.gz --strip-components 1 -C opensumi-release

# 解压opensumi-extension
sudo tar -zxf opensumi-extension.tar.gz --strip-components 1 -C opensumi-extension


echo 'ide.sh............end'