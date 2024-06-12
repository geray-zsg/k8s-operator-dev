
# 1.git操作
## 1.查看提交历史
```
git log
不传入任何参数的默认情况下，git log 会按时间先后顺序列出所有的提交，最近的更新排在最上面。这个命令会列出每个提交的 SHA-1 校验和、作者的名字和电子邮件地址、提交时间以及提交说明。

git log -p -2
-p 或 --patch ，它会显示每次提交所引入的差异（按 补丁 的格式输出）。 你也可以限制显示的日志条目数量，例如使用 -2 选项来只显示最近的两次提交

git log --stat
可以看到每次提交的简略统计信息。

git log --pretty=oneline
单行显示提交历史

git log --since=2.weeks
列出最近两周的所有提交
```

## 2.使用git tag给项目打标签
1.创建标签
```
git tag v1.0.0
```
默认标签是打在最新提交的commit上的。如果想给历史commit 打上标签，只需在后面加上 commit id 即可。`git tag v1.0.8 ba9f9e` 
2.上传标签
git push 并不会将 tag 推送到远程仓库服务器上，在创建完 tag 后我们需要手动推送 tag。
推送单个 tag：
```
git push origin v1.0.8
```
一次推送本地所有tag：
```
git push origin --tags
```
3.查看标签列表
```
git tag
```
通过 git tag -l "v1.0*" 查看 1.0.x 版本的tag

git tag -l 等同于 git tag --list；也可以使用 git tag --sort <key> 自定义 tag 排序

4.查看单个标签
使用 git show <tagname> 命令查看标签详细信息
```
git show v1.0
```
5.删除标签
使用 git tag -d <tagname> 删除本地仓库上的标签：
```
git tag -d v1.0
```
然后用 git push <remote> :refs/tags/<tagname> 更新远程仓库：
```
git push origin :refs/tags/v1.0.9
```
同步到本地
```
git fetch --prune --prune-tags
```
> 注意：git fetch --prune --prune-tags 会强制同步远程 tag 到本地，所以会导致本地新建的未提交到远程服务器的 tag 也会被删除。

6.给标签添加信息
上文提到的创建标签属于创建轻量标签，我们还可以在创建标签时通过-m <message>添加附加信息：
```
git tag v2.0.0 -m "version 2.0.0 released"
```
这样就对最新的提交添加了一个带附属信息的 tag。
添加多行信息可以添加多个 -m "<message>"
```
$ git tag v2.0.0 -m "version 2.0.0 released" -m "rebuild with react hooks" -m "support typescript"
```
这时候可以用 git tag -n<n> 查看 n 行的 tag 信息：
```
git tag -n3 // 查看三行的 tag 信息
```