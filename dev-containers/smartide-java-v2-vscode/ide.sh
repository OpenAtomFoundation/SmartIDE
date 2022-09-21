
###
 # @Author: kenan
 # @Date: 2022-07-07 14:24:14
 # @LastEditors: kenan
 # @LastEditTime: 2022-07-08 17:40:52
 # @FilePath: /smartide/dev-containers/smartide-base-v2-vscode/ide.sh
 # @Description: 
 # 
 # Copyright (c) 2022 by kenanlu@leansoftx.com, All Rights Reserved. 
### 
echo 'ide.sh............start'

if [ -d "./openvscode-images-amd64/" ];then
 sudo rm -rf openvscode-images-amd64
else
  echo 'openvscode-server............不存在'
fi

if [ -d "./openvscode-images-vmlc/" ];then
 sudo rm -rf openvscode-images-vmlc
else
  echo 'openvscode-server-vmlc.............不存在'
fi

if [ -d "./openvscode-images-arm64/" ];then
 sudo rm -rf openvscode-images-arm64
else
  echo 'openvscode-images-arm64.............不存在'
fi

if [ -d "./vsix/" ];then
 sudo rm -rf vsix
else
  echo 'vsix...........不存在'
fi

sudo mkdir openvscode-images-amd64 openvscode-images-vmlc openvscode-images-arm64 vsix vsix/extensions
sudo chmod -R 777 openvscode-images-amd64
sudo chmod -R 777 openvscode-images-vmlc
sudo chmod -R 777 openvscode-images-arm64
sudo chmod -R 777 vsix
sudo chmod -R 777 vsix/extensions


# 解压目录
sudo tar -zxf #{OpenVScodeServerVmlcFileName}#.tar.gz --strip-components 1 -C openvscode-images-amd64
sudo tar -zxf #{OpenVScodeServerVmlcFileName}#.tar.gz --strip-components 1 -C openvscode-images-vmlc
sudo tar -zxf #{OpenVScodeServerVmlcFileName}#-arm64.tar.gz --strip-components 1 -C openvscode-images-arm64

# 删除node   
sudo rm -rf ./openvscode-images-amd64/node
sudo rm -rf ./openvscode-images-vmlc/node
sudo rm -rf ./openvscode-images-arm64/node


# 解压插件
OPVSCODEVSIX=./vsix

for i in ./extensions/*.vsix;
    do
    sudo unzip $i "extension/*" -d $OPVSCODEVSIX/extensions/$(basename -s .vsix $i); \
    sudo mv $OPVSCODEVSIX/extensions/$(basename -s .vsix $i)/extension/* $OPVSCODEVSIX/extensions/$(basename -s .vsix $i); \
    sudo rm -rf $OPVSCODEVSIX/extensions/$(basename -s .vsix $i)/extension; \
    echo "$i........已复制"; \
    done

sudo \cp -rf ./vsix/extensions openvscode-images-amd64
sudo \cp -rf ./vsix/extensions  openvscode-images-vmlc
sudo \cp -rf ./vsix/extensions  openvscode-images-arm64

echo 'ide.sh............end'

