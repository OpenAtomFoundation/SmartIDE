---
title: "如何清理Git历史记录中的敏感信息"
linkTitle: "清理敏感信息"
date: 2022-07-22
weight: 20
description: >
  如果不小心在Git库引入了敏感信息，可能会造成严重的安全问题，比如泄露了密码，密钥或者token串等。以下介绍集中识别、清理和预防敏感信息提交的方法。
---

在合并拉取请求的时候选择Squash方法，我们可以有效的利用PR本身的功能避免敏感信息进入Git代码库，但是如果已经不小心引入了敏感信息，那么我们就需要方法能够识别并清理这些信息，否则可能会造成严重的安全问题。

以下介绍集中识别、清理和预防敏感信息提交的方法。

## 检测敏感信息

### 1. 使用git log搜索特定的敏感信息

如果你自己清楚需要定位的敏感信息内容，比如特定的关键字，特定的密码或者文件，那么可以采用通过执行以下指令来定位这些敏感信息。

```shell
## 直接搜索git history for string
git log -S <敏感词/密码内容等> --oneline
## 直接搜索git history for file
git log --all --full-history -- "**/thefile.*"
```

执行效果如下

![检测敏感信息](images/smartide-os007.png)

### 2. 使用Gitleaks定位敏感信息

Gitleaks是一款开源工具，可以自动根据预定义的规则扫描git库历史记录中可能的敏感信息，包括：密码，密钥，token等。Gitleaks的好处是无需提前知晓敏感信息内容，本身所提供的默认配置已经可以识别大多数可能的敏感信息内容，如果需要扩展也可以通过修改配置文件的方式（使用正则表达式）加入自己希望检测的pattern。

比如以下这个json就是gitleaks的一个可能包含了aws_secret的代码片段

```json
{
        "Description": "AWS",
        "StartLine": 37,
        "EndLine": 37,
        "StartColumn": 19,
        "EndColumn": 38,
        "Match": "\t\t\"aws_secret= \\\"AKIAIMNOJVGFDXXXE4OA\\\"\":          true,",
        "Secret": "AKIAIMNOJVGFDXXXE4OA",
        "File": "checks_test.go",
        "Commit": "ec2fc9d6cb0954fb3b57201cf6133c48d8ca0d29",
        "Entropy": 0,
        "Author": "zricethezav",
        "Email": "thisispublicanyways@gmail.com",
        "Date": "2018-01-28 17:39:00 -0500 -0500",
        "Message": "[update] entropy check",
        "Tags": [],
        "RuleID": "aws-access-token"
}
```

相关链接

- Gitleaks开源地址 https://github.com/zricethezav/gitleaks
- Azure DevOps 插件地址 https://marketplace.visualstudio.com/items?itemName=Foxholenl.Gitleaks


### 3. 使用detect secrets检查敏感信息

Detect secrets 和gitleaks的功能类似，不过detect secrets提供了更强大的插件机制，允许用户自行编写pattern识别机制。

相关链接：

- Detect secrets 开源地址 https://github.com/Yelp/detect-secrets 
- Code with Engineering Playbook 说明 https://microsoft.github.io/code-with-engineering-playbook/security/secret-detection/  


## 清理敏感信息

以上3种方式都可以有效的定位git历史记录中的敏感信息，但是定位以后的清理工作仍然是个麻烦的事情。因为git本身是一个链式数据结构，任何一个commit都会和前后的commit有链接关系，我们很难直接清理掉一个commit。而且敏感信息可能在历史中出现多次，清理掉一个可能会发现还会有其他，操作起来会也非常繁琐。

最有效的清理方式是将包含了敏感信息的文件先备份，然后从历史记录中彻底清理这个文件的记录，再把文件加回来。这样可以非常有效而且彻底的清理掉敏感信息。

当然这样做也有一个非常大的弊端就是这个文件会从之前的历史中消失，如果需要从历史时间点再次构建代码就会因为缺少文件而造成无法编译。同时，修改历史记录会造成对整个git历史记录重写（重新生成所有的commit记录），这样会造成很多其他关联系统的链接失效（比如在issue中关联了commitID）。但是相对于遗留敏感信息在历史记录中，类似的问题可能就不那么重要了。

以下脚本可以将生成一个包含了所有已经从代码库中删除的文件的清单，然后根据这个清单彻底清理涉及这个文件的历史记录。

```shell
## 以下脚本只能在Linux上运行
## 生成已删除文件清单 deleted.txt
git log --diff-filter=D --summary <起始点的CommitID>..HEAD | egrep -o '*[[:alnum:]]*(/[[:alnum:].]*)+$' > deleted.txt

## 清理历史记录
for del in `cat deleted.txt`
do
    git filter-branch --index-filter "git rm --cached --ignore-unmatch $del" --prune-empty -- --all
    # The following seems to be necessary every time
    # because otherwise git won't overwrite refs/original
    git reset --hard
    git for-each-ref --format="%(refname)" refs/original/ | xargs -n 1 git update-ref -d
    git reflog expire --expire=now --all
    git gc --aggressive --prune=now
done;
```

## 防止提交敏感信息

以上方法都只能起到亡羊补牢的作用，而且清理操作会造成破坏性的结果，对保持项目历史记录延续性非常不利。因此最有效的方式是从根本上防止问题的发生。

Gitleaks和detect secret这两个工具都可以设置成 git pre-commit hook（预提交检查），这样在开发者试图提交带有敏感信息的文件的时候就会直接拒绝commit的创建，从而从根本上解决敏感信息的问题。

具体操作方式可以参考：

- 设置Gitleaks作为预提交检查工具 https://github.com/zricethezav/gitleaks#pre-commit-hook 
- 设置detect secret作为预提交检查工具 https://microsoft.github.io/code-with-engineering-playbook/security/secret-detection/#detect-secrets-pre-commit-hook

