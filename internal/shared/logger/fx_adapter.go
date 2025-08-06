package logger

import (
	"fmt"
	"strings"

	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

// FxAdapter 适配器，将 Fx 事件转换为统一的日志格式
// 遵循 Zap 最佳实践：
// - Info: 应用生命周期的关键事件 (Started, Stopped, 错误)
// - Debug: 内部依赖注入过程 (Provided, Invoking等)
// - Error: 所有失败的操作
type FxAdapter struct {
	logger Logger
}

// NewFxAdapter 创建 Fx 日志适配器
func NewFxAdapter(logger Logger) fxevent.Logger {
	return &FxAdapter{
		logger: logger,
	}
}

// LogEvent 处理 Fx 事件并转换为统一日志格式
func (l *FxAdapter) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		// 只在debug级别记录生命周期钩子执行，避免启动噪音
		l.logger.Debug("Fx lifecycle hook executing",
			zap.String("hook", "OnStart"),
			zap.String("function", e.FunctionName),
			zap.String("caller", e.CallerName))

	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			l.logger.Error("Fx lifecycle hook failed",
				zap.String("hook", "OnStart"),
				zap.String("function", e.FunctionName),
				zap.String("caller", e.CallerName),
				zap.String("runtime", e.Runtime.String()),
				zap.Error(e.Err))
		} else {
			// 成功的生命周期钩子在debug级别记录
			l.logger.Debug("Fx lifecycle hook executed successfully",
				zap.String("hook", "OnStart"),
				zap.String("function", e.FunctionName),
				zap.String("caller", e.CallerName),
				zap.String("runtime", e.Runtime.String()))
		}

	case *fxevent.OnStopExecuting:
		l.logger.Debug("Fx lifecycle hook executing",
			zap.String("hook", "OnStop"),
			zap.String("function", e.FunctionName),
			zap.String("caller", e.CallerName))

	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			l.logger.Error("Fx lifecycle hook failed",
				zap.String("hook", "OnStop"),
				zap.String("function", e.FunctionName),
				zap.String("caller", e.CallerName),
				zap.String("runtime", e.Runtime.String()),
				zap.Error(e.Err))
		} else {
			l.logger.Debug("Fx lifecycle hook executed successfully",
				zap.String("hook", "OnStop"),
				zap.String("function", e.FunctionName),
				zap.String("caller", e.CallerName),
				zap.String("runtime", e.Runtime.String()))
		}

	case *fxevent.Supplied:
		l.logger.Debug("Fx dependency supplied",
			zap.String("type", e.TypeName),
			zap.String("module", e.ModuleName))

	case *fxevent.Provided:
		typeName := "<unknown>"
		if len(e.OutputTypeNames) > 0 {
			typeName = e.OutputTypeNames[0]
		}
		l.logger.Debug("Fx dependency provided",
			zap.String("constructor", cleanConstructorName(e.ConstructorName)),
			zap.String("type", typeName),
			zap.String("module", e.ModuleName))

	case *fxevent.Decorated:
		typeName := "<unknown>"
		if len(e.OutputTypeNames) > 0 {
			typeName = e.OutputTypeNames[0]
		}
		l.logger.Debug("Fx dependency decorated",
			zap.String("decorator", cleanConstructorName(e.DecoratorName)),
			zap.String("type", typeName),
			zap.String("module", e.ModuleName))

	case *fxevent.Invoking:
		// 只在debug级别记录函数调用，避免启动时的日志噪音
		l.logger.Debug("Fx invoking function",
			zap.String("function", cleanConstructorName(e.FunctionName)),
			zap.String("module", e.ModuleName))

	case *fxevent.Invoked:
		if e.Err != nil {
			l.logger.Error("Fx function invocation failed",
				zap.String("function", cleanConstructorName(e.FunctionName)),
				zap.String("module", e.ModuleName),
				zap.Error(e.Err))
		} else {
			l.logger.Debug("Fx function invoked successfully",
				zap.String("function", cleanConstructorName(e.FunctionName)),
				zap.String("module", e.ModuleName))
		}

	case *fxevent.Stopping:
		l.logger.Info("Fx application stopping",
			zap.String("signal", e.Signal.String()))

	case *fxevent.Stopped:
		if e.Err != nil {
			l.logger.Error("Fx application stop failed", zap.Error(e.Err))
		} else {
			l.logger.Info("Fx application stopped successfully")
		}

	case *fxevent.RollingBack:
		l.logger.Warn("Fx rolling back startup", zap.Error(e.StartErr))

	case *fxevent.RolledBack:
		if e.Err != nil {
			l.logger.Error("Fx rollback failed", zap.Error(e.Err))
		} else {
			l.logger.Info("Fx rollback successful")
		}

	case *fxevent.Started:
		// 移动到debug级别，避免与业务应用启动日志重复
		l.logger.Debug("Fx dependency injection framework started")

	case *fxevent.LoggerInitialized:
		l.logger.Debug("Fx custom logger initialized",
			zap.String("constructor", cleanConstructorName(e.ConstructorName)))

	default:
		// 对于未知的事件类型，使用通用日志
		l.logger.Debug("Fx event", zap.String("event_type", fmt.Sprintf("%T", event)))
	}
}

// cleanConstructorName 清理构造函数名称，移除包路径使其更简洁
func cleanConstructorName(name string) string {
	// 移除包路径，只保留函数名
	if lastSlash := strings.LastIndex(name, "/"); lastSlash != -1 {
		name = name[lastSlash+1:]
	}

	// 如果有模块名前缀，也简化
	if dot := strings.Index(name, "."); dot != -1 && !strings.Contains(name[:dot], "(") {
		// 保留模块名，但简化包路径
		parts := strings.Split(name, ".")
		if len(parts) >= 2 {
			return parts[len(parts)-2] + "." + parts[len(parts)-1]
		}
	}

	return name
}
