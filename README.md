# go_modifymeta
工具说明：
  1. 此工具用于将对象存储桶中 mp4类型的对象的content-type值修改为 video/mp4. 
  2. 更广泛的应用可以基于此进行简单修改，就可以实现更多不同的元数据值的批量大并发修改
  
使用说明：
  
  1. 执行命令： ./go_modifymeta  -config=config.dat -job=20 
  2. 参数说明： config做配置参数带入，job定义了任务执行并发度；config,dat中 配置桶名，区域，AK/SK，自定义域名配置信息。
