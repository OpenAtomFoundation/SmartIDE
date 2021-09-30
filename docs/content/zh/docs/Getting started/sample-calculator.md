---
title: "示例：计算器应用"
linkTitle: "示例：计算器应用"
weight: 10
date: 2021-09-29
description: >
  本应用运行状态为网页中的计算器，使用node.js创建，并包含了试用mocha的单元测试代码，如下图：
  
  ![](images/calculator-ui.png)

  代码中使用node.js代码提供了REST APIs，其中提供各种数学计算功能单元。
  使用mocah编写的测试代码可以完成所有以上API内部运算运算逻辑的验证，最终使用 mocha-junit-reports 来生成XML格式的测试结果文件以便 Azure DevOps 可以读取测试结果提供DevOps流水线的测试集成。
---

## 场景

- SmartIDE本地运行，使用WebIDE方式调试
- 使用本地VScode链接smartIDE开发容器调试

## 整体说明

![](images/process-all.png)

## 先决条件

安装SmartIDE，参考链接: https://smartide.dev/zh/docs/getting-started/install/

Demo源码获取地址，https://github.com/idcf-boat-house/boathouse-calculator/blob/master/README.md

##  场景1.SmartIDE本地运行，使用WebIDE方式调试



##  场景2.使用本地VScode链接smartIDE开发容器调试

