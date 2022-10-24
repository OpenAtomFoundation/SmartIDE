#################################################
# SmartIDE Developer Container Image
# Licensed under GPL v3.0
# Copyright (C) leansoftX.com
#################################################

# ARG LLVM=15
# ARG CMAKE_VERSION=3.24.0

FROM registry.cn-hangzhou.aliyuncs.com/smartide/smartide-node-v2:latest
# FROM registry.cn-hangzhou.aliyuncs.com/smartide/smartide-node-v2-vmlc:latest
USER root

RUN apt-get update && \
    apt-get -y install --no-install-recommends gcc cmake clang build-essential libssl-dev \
    cppcheck valgrind lldb llvm gdb ninja-build \
    libjansson-dev libsnappy-dev liblzma-dev libz-dev pkg-config && \
    apt-get clean && \
    apt-get autoremove -y && \
    rm -rf /tmp/* /var/lib/apt/lists/* /var/tmp/*

# Install clangd and clang-tidy from the public LLVM PPA (nightly build / development version)
# And also the GDB debugger from the Ubuntu repos
# ARG LLVM=15

# RUN echo "LLVM:$LLVM" && \
#     wget https://apt.llvm.org/llvm.sh && \
#     chmod +x llvm.sh && \
#     ./llvm.sh $LLVM && \
#     apt-get clean && \
#     apt-get autoremove -y && \
#     rm -rf /tmp/* /var/lib/apt/lists/* /var/tmp/*

# ARG CMAKE_VERSION=3.24.0
# # Install latest stable CMake
# RUN echo "CMAKE_VERSION:$CMAKE_VERSION" && \
#     wget "https://github.com/Kitware/CMake/releases/download/v$CMAKE_VERSION/cmake-$CMAKE_VERSION-Linux-x86_64.sh" && \
#     chmod a+x cmake-$CMAKE_VERSION-Linux-x86_64.sh && \
#     ./cmake-$CMAKE_VERSION-Linux-x86_64.sh --prefix=/usr/ --skip-license && \
#     rm cmake-$CMAKE_VERSION-Linux-x86_64.sh && \
#     apt-get clean && \
#     apt-get autoremove -y && \
#     rm -rf /tmp/* /var/lib/apt/lists/* /var/tmp/*