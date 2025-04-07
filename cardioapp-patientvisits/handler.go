package function

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"handler/function/config"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"firebase.google.com/go/v4/messaging"
	"github.com/appleboy/go-fcm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/spf13/cast"
)

const (
	appId = "P-JV2nVIRUtgyPO5xRNeYll2mT4F5QG4bS"
)
