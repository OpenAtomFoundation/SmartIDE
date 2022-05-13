
echo 'ide.sh............start'

if [ -d "./openvscode-images/" ];then
 sudo rm -rf openvscode-images
else
  echo 'openvscode-server............不存在'
fi

if [ -d "./vsix/" ];then
 sudo rm -rf vsix
else
  echo 'vsix...........不存在'
fi

sudo mkdir openvscode-images vsix vsix/extensions
sudo chmod -R 777 openvscode-images
sudo chmod -R 777 vsix
sudo chmod -R 777 vsix/extensions


# 解压目录
sudo tar -zxf #{OpenVScodeServerFileName}#.tar.gz --strip-components 1 -C openvscode-images

# 删除node   
sudo rm -rf ./openvscode-images/node

# 删除server.sh
# sudo rm -rf ./openvscode-images/server.sh
# 复制server.sh
# sudo cp server.sh ./openvscode-images/
# sudo chmod +x ./openvscode-images/server.sh;


# 解压插件
# OPVSCODEVSIX=./vsix

# for i in ./extensions/*.vsix;
#     do
#     sudo unzip $i "extension/*" -d $OPVSCODEVSIX/extensions/$(basename -s .vsix $i); \
#     sudo mv $OPVSCODEVSIX/extensions/$(basename -s .vsix $i)/extension/* $OPVSCODEVSIX/extensions/$(basename -s .vsix $i); \
#     sudo rm -rf $OPVSCODEVSIX/extensions/$(basename -s .vsix $i)/extension; \
#     echo "$i........已复制"; \
#     done

# sudo \cp -rf ./vsix/extensions openvscode-images

echo 'ide.sh............end'

