package fingprints

import (
    "github.com/yhy0/ChYing/pkg/Jie/fingprints/framework"
    "github.com/yhy0/ChYing/pkg/Jie/fingprints/os"
    "github.com/yhy0/ChYing/pkg/Jie/fingprints/programing"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

var ProgramingPlugins = [...]Plugin{
    &programing.AsciiDocPlugin{},
    &programing.AspPlugin{},
    &programing.ErlangPlugin{},
    &programing.JavaPlugin{},
    &programing.LuaPlugin{},
    &programing.PerlPlugin{},
    &programing.PHPPlugin{},
    &programing.PythonPlugin{},
    &programing.RubyPlugin{},
}

var OsPlugins = [...]Plugin{
    &os.CentOSPlugin{},
    &os.DarwinPlugin{},
    &os.DebianPlugin{},
    &os.FedoraPlugin{},
    &os.FreeBSDPlugin{},
    &os.GentooPlugin{},
    &os.RedHatPlugin{},
    &os.ScientificPlugin{},
    &os.SunOSPlugin{},
    &os.SUSEPlugin{},
    &os.UbuntuPlugin{},
    &os.UNIXPlugin{},
    &os.WindowsCEPlugin{},
    &os.WindowsServerPlugin{},
}

var FrameworkPlugins = [...]Plugin{
    &framework.ApacheJackrabbitPlugin{},
    &framework.AspMvcPlugin{},
    &framework.CakePHPPlugin{},
    &framework.CherryPyPlugin{},
    &framework.CodeIgniterPlugin{},
    &framework.DancerPlugin{},
    &framework.DjangoPlugin{},
    &framework.FlaskPlugin{},
    &framework.FuelPHPPlugin{},
    &framework.GrailsPlugin{},
    &framework.GrailsPlugin{},
    &framework.HordePlugin{},
    &framework.KarrigellPlugin{},
    &framework.LaravelPlugin{},
    &framework.NettePlugin{},
    &framework.PhalconPlugin{},
    &framework.PlayPlugin{},
    &framework.RailsPlugin{},
    &framework.SeagullPlugin{},
    &framework.SpringPlugin{},
    &framework.SymfonyPlugin{},
    &framework.Web2PyPlugin{},
    &framework.YiiPlugin{},
    &framework.ZendPlugin{},
    
    &framework.BeegoPlugin{},
}
