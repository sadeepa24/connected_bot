package connected

import (
	"context"
	"errors"

	"github.com/sadeepa24/connected_bot/botapi"
	"github.com/sadeepa24/connected_bot/controller"
	"github.com/sadeepa24/connected_bot/db"
	"github.com/sadeepa24/connected_bot/parser"
	"github.com/sadeepa24/connected_bot/server"
	"github.com/sadeepa24/connected_bot/service"
	"github.com/sadeepa24/connected_bot/watchman"
	"go.uber.org/zap"
)

type ConnectedBot struct {
	Ctx context.Context

	db     *db.Database //only for parsing to watchman and controller
	Botapi botapi.BotAPI
	logger *zap.Logger

	Webhookserver *server.Webhookserver
	Parser        parser.Parserwrap
	Ctrl          *controller.Controller
	Watchman      *watchman.Watchman
	Services      []service.Service

	msgstore *botapi.MessageStore
}

func New(options Botoptions) (*ConnectedBot, error) {
	newdb := db.New(options.Ctx, options.Logger, options.Dbpath)
	var (
		//sbox sbox.Sboxcontroller
		newbotapi botapi.BotAPI
	)
	msgstore, err := botapi.NewMessageStore(options.TemplatesPath)
	if err != nil {
		return nil, err
	}
	newbotapi = botapi.NewBot(options.Ctx, options.Bottoken, options.Botmainurl, msgstore)
	//sbox = singapi.New(options.Sboxconf)  // TODO: should add argument ctx, options,

	//testings
	//sbox = fortest.Newtestsbox()
	//newbotapi = fortest.NewTESTBOTAPI()

	ctrl, err := controller.New(options.Ctx, newdb, options.Logger, options.Metadata, newbotapi, options.Sboxoption)
	if err != nil {
		return nil, err
	}
	watchman := watchman.New(options.Ctx, ctrl, newbotapi, newdb, options.Watchman, options.Logger, msgstore)
	srvs, err := service.GetallService(options.Ctx, options.Logger, ctrl, newbotapi, msgstore)
	if err != nil {
		return nil, err
	}
	parser := parser.New(options.Ctx, ctrl, srvs, newbotapi, options.Logger)

	return &ConnectedBot{
		db:            newdb,
		logger:        options.Logger,
		Botapi:        newbotapi,
		Ctrl:          ctrl,
		Webhookserver: server.New(options.Ctx, options.WebHookServerOption, parser, options.Logger),
		Parser:        parser,
		Services:      srvs,
		Watchman:      watchman,
		msgstore: msgstore,
	}, nil

}

func (c *ConnectedBot) Start() error {
	var err error
	if err = c.db.InitDb(); err != nil { // initing db so that other can intract with it (create, update, quary )
		return err
	}
	c.logger.Debug("database inited")
	if err = c.Ctrl.Init(); err != nil {
		return err
	}
	c.logger.Debug("controller inited")

	if err = c.Watchman.Start(); err != nil {
		return err
	}
	c.logger.Debug("watchman started")

	if err = c.msgstore.Init(c.Botapi, c.Ctrl.SudoAdmin, c.logger); err != nil {
		return err
	}
	c.logger.Debug("messagestore inited")
	
	for _, service := range c.Services {
		c.logger.Info("service started " + service.Name())
		if err = service.Init(); err != nil {
			return err
		}

	}
	if err = c.Parser.Init(); err != nil {
		return err
	}
	c.logger.Debug("parser inited")
	if c.Webhookserver == nil {
		return errors.New("webhook server not found")
	}
	webhookerrRecive := make(chan error)
	go c.Webhookserver.Start(c.Botapi, webhookerrRecive)
	err = <-webhookerrRecive
	c.logger.Info("ARE YOU HAPPY NOW 😁😎")

	return err
}

func (c *ConnectedBot) Close() error {
	var err error
	err = errors.Join(c.Webhookserver.Close(), err)  // close webhookserver so that new update does not recive
	err = errors.Join(c.Parser.Stop(), err)          // stops all ongoing updatectx
	err = errors.Join(c.Watchman.Close(), err) // close watchman
	err = errors.Join(c.Ctrl.Close(), err)
	err = errors.Join(c.db.Close(), err)

	if err != nil {
		c.logger.Error("closing error detected ", zap.Error(err))
	} else {
		c.logger.Info("Everything Closed Successfully. Program WIll exit Safaly")
	}

	return err

}
