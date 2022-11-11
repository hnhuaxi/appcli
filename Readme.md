# Appcli 一种基于 YAML 配置生成命令行应用的工具

## 概念
Console cli 命令行交互式的程序一直以来都是一种重要的生产力方式，长期以来在系统级、编程、调试、运维方面有着其它交互模式不可替代的作用，它有着便捷性，定制性，与自动化主要特点，可以轻松编写脚本、原生的字符流管道处理能力，是构成自动化任务中的重要组成，在今天 CI, DI, DevOps 处于核心的位置，可以说今天任何一款复杂的云服务系统如果没有了 CLI 工具，那就是缺乏了定制化，自动化的部分，也可以说整个系统是不完整的。

但是编写 CLI 的工作也不是一件轻松的工作，尤其是要有跨平台的能力与兼容性，需要一种能够快速实现控制台工具生成的开发工具，Appcli 以解决这个目的而生。

通过 YAML 文档生成一个程序，代码编写的过程中有很多的重复，甚至框架都是通用的，我们要关注的实际上是特定的领域，把通用的部分提炼出来编写代码，为特性编写配置来实现动态化，这就是 Appcli 的工作方式，不用发明一种 DSL 只是基于配置，你不用学习一种新语法，所以你的掌握速度会更快，学习使用配置及查表能更好的生成你想要的程序。


