package ecs

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/session-manager-plugin/src/datachannel"
	sessionlog "github.com/aws/session-manager-plugin/src/log"
	"github.com/aws/session-manager-plugin/src/sessionmanagerplugin/session"
	"github.com/aws/session-manager-plugin/src/sessionmanagerplugin/session/portsession"
	"github.com/twinj/uuid"

	f "github.com/danielunderwood/portctl/forwarding"
)

func handleError(err error, log *Logger) {
	if err != nil {
		log.Critical(err)
	}
}

// Wrapper type for log.Logger so session-manager-plugin can use it
type Logger struct {
	inner *log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		inner: log.Default(),
	}
}

// Tracef formats message according to format specifier
// and writes to log with level Trace.
func (l *Logger) Tracef(format string, params ...interface{}) {
	l.inner.Printf(format, params...)
}

// Debugf formats message according to format specifier
// and writes to log with level Debug.
func (l *Logger) Debugf(format string, params ...interface{}) {
	l.inner.Printf(format, params...)
}

// Infof formats message according to format specifier
// and writes to log with level Info.
func (l *Logger) Infof(format string, params ...interface{}) {
	l.inner.Printf(format, params...)
}

// Warnf formats message according to format specifier
// and writes to log with level Warn.
func (l *Logger) Warnf(format string, params ...interface{}) error {
	l.inner.Printf(format, params...)
	return nil
}

// Errorf formats message according to format specifier
// and writes to log with level Error.
func (l *Logger) Errorf(format string, params ...interface{}) error {
	l.inner.Panicf(format, params...)
	return nil
}

// Criticalf formats message according to format specifier
// and writes to log with level Critical.
func (l *Logger) Criticalf(format string, params ...interface{}) error {
	l.inner.Fatalf(format, params...)
	return nil
}

// Trace formats message using the default formats for its operands
// and writes to log with level Trace.
func (l *Logger) Trace(v ...interface{}) {
	l.inner.Print(v...)
}

// Debug formats message using the default formats for its operands
// and writes to log with level Debug.
func (l *Logger) Debug(v ...interface{}) {
	l.inner.Print(v...)
}

// Info formats message using the default formats for its operands
// and writes to log with level Info.
func (l *Logger) Info(v ...interface{}) {
	l.inner.Print(v...)
}

// Warn formats message using the default formats for its operands
// and writes to log with level Warn.
func (l *Logger) Warn(v ...interface{}) error {
	l.inner.Print(v...)
	return nil
}

// Error formats message using the default formats for its operands
// and writes to log with level Error.
func (l *Logger) Error(v ...interface{}) error {
	l.inner.Print(v...)
	return nil
}

// Critical formats message using the default formats for its operands
// and writes to log with level Critical.
func (l *Logger) Critical(v ...interface{}) error {
	l.inner.Print(v...)
	return nil
}

// Flush flushes all the messages in the logger.
func (l *Logger) Flush() {

}

// Close flushes all the messages in the logger and closes it. The logger cannot be used after this operation.
func (l *Logger) Close() {

}

func (l *Logger) WithContext(context ...string) (contextLogger sessionlog.T) { return l }

// Forward a port from ECS by finding the target corresponding to ECS cluster
// and container name
func ForwardECS(forwardConfig f.ECSForwardRule) {
	log.Println("Forwarding ECS with dynamic settings")
	config, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	client := ecs.NewFromConfig(config)
	tasksInput := ecs.ListTasksInput{
		Cluster: aws.String(forwardConfig.Cluster),
	}
	tasks, err := client.ListTasks(context.TODO(), &tasksInput)
	if err != nil {
		log.Fatalf("Error listing tasks: %v", err)
	}

	describeTasksInput := ecs.DescribeTasksInput{
		Cluster: aws.String(forwardConfig.Cluster),
		Tasks:   tasks.TaskArns,
	}
	describedTasks, err := client.DescribeTasks(context.TODO(), &describeTasksInput)
	if err != nil {
		log.Fatalf("Error describing tasks: %v", err)
	}

	var runtimeId string
	var taskId string
	for _, task := range describedTasks.Tasks {
		for _, container := range task.Containers {
			// TODO: It's possible that containers across different tasks are
			// named the same thing. This should be handled.
			if *container.Name == forwardConfig.Container {
				runtimeId = *container.RuntimeId
				taskParts := strings.Split(*task.TaskArn, "/")
				taskId = taskParts[len(taskParts)-1]
				break
			}
		}
	}
	if runtimeId == "" {
		log.Fatalf("Could not find container %s in cluster %s", forwardConfig.Container, forwardConfig.Cluster)
	}

	log.Printf("Found runtime ID %s", runtimeId)

	target := fmt.Sprintf("ecs:%s_%s_%s", forwardConfig.Cluster, taskId, runtimeId)
	forwardECSTarget(forwardConfig, target)
}

func forwardECSTarget(forwardConfig f.ECSForwardRule, target string) {
	logger := NewLogger()
	logger.Info("Starting session manager forwarding")

	logger.Infof("Have temp directory %s", os.TempDir())
	// return

	cfg, err := config.LoadDefaultConfig(context.TODO())
	handleError(err, logger)

	stsClient := sts.NewFromConfig(cfg)
	callerIdentity, err := stsClient.GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})
	handleError(err, logger)
	logger.Infof("Have identity %s", *callerIdentity.Arn)

	ssmClient := ssm.NewFromConfig(cfg)
	// TODO: This could be specified separately or "dynamic" would choose a
	// random unused port
	localPort := strconv.FormatUint(forwardConfig.Base.Port, 10)
	remotePort := strconv.FormatUint(forwardConfig.Base.Port, 10)
	params := map[string][]string{"host": {"localhost"}, "portNumber": {remotePort}, "localPortNumber": {localPort}}
	logger.Infof("Params: %v", params)
	sessionParams, err := ssmClient.StartSession(context.TODO(), &ssm.StartSessionInput{
		// TODO: Fetch Dynamically
		Target:       aws.String(target),
		DocumentName: aws.String("AWS-StartPortForwardingSessionToRemoteHost"),
		Parameters:   params,
	})
	handleError(err, logger)

	// This doesn't seem to be in the plugin code, but there's a segfault without
	// it. Maybe it's how plugins are initialized externally?
	session.SessionRegistry = make(map[string]session.ISessionPlugin)
	session.Register(&portsession.PortSession{})
	// Following https://github.com/aws/session-manager-plugin/blob/mainline/src/sessionmanagerplugin/session/session.go#L200C3-L200C10
	uuid.SwitchFormat(uuid.CleanHyphen)
	clientId := uuid.NewV4().String()
	var s session.Session
	s.SessionId = *sessionParams.SessionId
	s.TokenValue = *sessionParams.TokenValue
	s.StreamUrl = *sessionParams.StreamUrl
	s.ClientId = clientId
	s.TargetId = target
	s.DataChannel = &datachannel.DataChannel{}

	// NOTE: This needs TMPDIR=/tmp to work on mac
	err = s.Execute(logger)
	handleError(err, logger)

}
