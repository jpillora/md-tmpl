### Usage

<!--tmpl,chomp,code=plain:md-tmpl --help -->
``` plain 

  Usage: md-tmpl [options] files...

  Markdown template will look for 'tmpl' HTML comments in the give files.
  Templates must be in the format:
      <!--tmpl: my-command --><!--/tmpl-->

  In this case, 'my-command' would be executed via bash and the output would
  be inserted in between the start and end 'tmpl' tags.

  Options:
  --preview, -p  Enables preview mode. Displays all commands encountered.
  --write, -w    Write file instead of printing to standard out
  --help, -h

```
<!--/tmpl-->

### Stuff

...
