package wraith

import (
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

type Page struct {
	rod *rod.Page
}

func (p *Page) Navigate(url string) error {
	return p.rod.Navigate(url)
}

func (p *Page) WaitLoad() error {
	return p.rod.WaitLoad()
}

func (p *Page) WaitIdle(timeout time.Duration) error {
	return p.rod.WaitIdle(timeout)
}

func (p *Page) WaitForSelector(selector string, timeout time.Duration) error {
	el, err := p.rod.Timeout(timeout).Element(selector)
	if err != nil {
		return err
	}
	return el.WaitVisible()
}

func (p *Page) Click(selector string) error {
	el, err := p.rod.Element(selector)
	if err != nil {
		return err
	}
	return el.Click(proto.InputMouseButtonLeft, 1)
}

func (p *Page) Fill(selector, text string) error {
	el, err := p.rod.Element(selector)
	if err != nil {
		return err
	}
	return el.Input(text)
}

func (p *Page) Text(selector string) (string, error) {
	el, err := p.rod.Element(selector)
	if err != nil {
		return "", err
	}
	return el.Text()
}

func (p *Page) HTML() (string, error) {
	return p.rod.HTML()
}

func (p *Page) Title() (string, error) {
	info, err := p.rod.Info()
	if err != nil {
		return "", err
	}
	return info.Title, nil
}

func (p *Page) URL() (string, error) {
	info, err := p.rod.Info()
	if err != nil {
		return "", err
	}
	return info.URL, nil
}

func (p *Page) Eval(js string) (*proto.RuntimeRemoteObject, error) {
	return p.rod.Eval(js)
}

func (p *Page) Screenshot(path string) error {
	img, err := p.rod.Screenshot(true, nil)
	if err != nil {
		return err
	}
	return writeFile(path, img)
}

func (p *Page) Cookies() ([]*proto.NetworkCookie, error) {
	return p.rod.Cookies(nil)
}

func (p *Page) Close() error {
	return p.rod.Close()
}

func (p *Page) RodPage() *rod.Page { return p.rod }
