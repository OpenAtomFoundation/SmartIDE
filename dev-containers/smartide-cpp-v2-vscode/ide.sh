#################################################
# SmartIDE Developer Container Image
# Licensed under GPL v3.0
# Copyright (C) leansoftX.com
#################################################

echo 'ide.sh............start'

if [ -d "./openvscode-images/" ];then
 sudo rm -rf openvscode-images
else
  echo 'openvscode-server............不存在'
fi

if [ -d "./openvscode-images-vmlc/" ];then
 sudo rm -rf openvscode-images-vmlc
else
  echo 'openvscode-server-vmlc.............不存在'
fi

if [ -d "./vsix/" ];then
 sudo rm -rf vsix
else
  echo 'vsix...........不存在'
fi

sudo mkdir openvscode-images openvscode-images-vmlc vsix vsix/extensions
sudo chmod -R 777 openvscode-images
sudo chmod -R 777 openvscode-images-vmlc
sudo chmod -R 777 vsix
sudo chmod -R 777 vsix/extensions


# 解压目录
sudo tar -zxf #{OpenVScodeServerVmlcFileName}#.tar.gz --strip-components 1 -C openvscode-images
sudo tar -zxf #{OpenVScodeServerVmlcFileName}#.tar.gz --strip-components 1 -C openvscode-images-vmlc

# 删除node   
sudo rm -rf ./openvscode-images/node
sudo rm -rf ./openvscode-images-vmlc/node

# 解压插件
OPVSCODEVSIX=./vsix

for i in ./extensions/*.vsix;
    do
    sudo unzip $i "extension/*" -d $OPVSCODEVSIX/extensions/$(basename -s .vsix $i); \
    sudo mv $OPVSCODEVSIX/extensions/$(basename -s .vsix $i)/extension/* $OPVSCODEVSIX/extensions/$(basename -s .vsix $i); \
    sudo rm -rf $OPVSCODEVSIX/extensions/$(basename -s .vsix $i)/extension; \
    echo "$i........已复制"; \
    done

sudo \cp -rf ./vsix/extensions openvscode-images
sudo \cp -rf ./vsix/extensions openvscode-images-vmlc

echo 'ide.sh............end'

echo "--------openvscode-images目录文件-------"
ls -a ./openvscode-images
echo "--------openvscode-images-vmlc目录文件-------"
ls -a ./openvscode-images-vmlc
echo "--------/vsix/extensions目录文件-------"
ls -a ./vsix/extensions